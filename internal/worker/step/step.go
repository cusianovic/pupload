package step

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gabriel-vasile/mimetype"
	mimetypes "github.com/pupload/pupload/internal/mimetype"
	"github.com/pupload/pupload/internal/models"
	"github.com/pupload/pupload/internal/resources"
	"github.com/pupload/pupload/internal/syncplane"
	"github.com/pupload/pupload/internal/worker/container"

	"github.com/google/uuid"
)

type StepService struct {
	SyncLayer      syncplane.SyncLayer
	CS             *container.ContainerService
	ResourceManger *resources.ResourceManager

	mu sync.Mutex
}

func CreateStepService(cs *container.ContainerService, s syncplane.SyncLayer, rm *resources.ResourceManager) (StepService, error) {

	err := s.UpdateSubscribedQueues(rm.GetValidTierMap())
	if err != nil {
		return StepService{}, err
	}

	return StepService{
		CS:             cs,
		SyncLayer:      s,
		ResourceManger: rm,
	}, nil
}

func (ss *StepService) addEnvFlagMap(m map[string]string, task models.Task, step models.Step) error {

	flags, err := ss.getFlags(task, step)
	if err != nil {
		return err
	}

	// TODO: ensure this can't be escaped
	for key, val := range flags {
		m[key] = val
	}

	return nil
}

type preparedIO struct {
	name      string
	base_path string
	path      string
	filename  string
	url       string
}

func (ss *StepService) prepareIO(inputs, outputs map[string]string, task models.Task, basePath string) ([]preparedIO, []preparedIO, error) {
	in := make([]preparedIO, 0, len(inputs))
	out := make([]preparedIO, 0, len(outputs))

	for _, inputDef := range task.Inputs {
		inputURL, ok := inputs[inputDef.Name]
		if !ok {
			switch inputDef.Required {
			case true:
				return nil, nil, fmt.Errorf("PrepareInputs: step missing required input %s", inputDef.Name)
			case false:
				continue
			}
		}

		typeSet, err := mimetypes.CreateMimeSet(inputDef.Type)
		if err != nil {
			return nil, nil, fmt.Errorf("PrepareInputs: error creating mimeset: %w", err)
		}

		ext, err := ss.validateInput(inputURL, *typeSet)
		if err != nil {
			return nil, nil, fmt.Errorf("PrepareInputs: error validating inputs: %w", err)
		}

		path, filename := ss.getPath(basePath, ext)

		in = append(in, preparedIO{
			name:      inputDef.Name,
			url:       inputURL,
			base_path: basePath,
			path:      path,
			filename:  filename,
		})

	}

	for _, outputDef := range task.Outputs {
		outputURL, ok := outputs[outputDef.Name]
		if !ok {
			return nil, nil, fmt.Errorf("no output URL for output %s", outputDef.Name)
		}

		extension := ss.getOutputExtension(outputDef.Type)
		path, filename := ss.getPath(basePath, extension)

		out = append(out, preparedIO{
			url:       outputURL,
			name:      outputDef.Name,
			base_path: basePath,
			path:      path,
			filename:  filename,
		})
	}

	return in, out, nil
}

func (ss *StepService) addIOToEnvMap(env map[string]string, prepped []preparedIO) {
	for _, prep := range prepped {
		env[prep.name] = prep.path
	}
}

func (ss *StepService) getOutputExtension(types []models.MimeType) string {
	if len(types) != 1 {
		return ""
	}

	t := types[0]
	return mimetypes.GetExtensionFromMime(t)
}

// Validates a given uploaded file against the qualified allowed mime types.
// Returns the appoprriate file extension
func (ss *StepService) validateInput(url string, mimeSet mimetypes.MimeSet) (ext string, err error) {

	resp, err := http.Get(url)

	if err != nil {
		return "", fmt.Errorf("error getting content from %s", url)
	}

	defer resp.Body.Close()

	// Try the Content-Type header first (S3 and most storage backends set this).
	// Strip parameters (e.g. "video/mp4; codecs=avc1" -> "video/mp4").
	mime := stripMimeParams(resp.Header.Get("Content-Type"))

	// Fall back to byte sniffing if the header is missing or generic.
	if mime == "" || mime == "application/octet-stream" || mime == "binary/octet-stream" {
		mimeBytes := make([]byte, 512)
		io.ReadFull(resp.Body, mimeBytes)
		mt := mimetype.Detect(mimeBytes)
		mime = stripMimeParams(mt.String())
	}

	if !mimeSet.Contains(models.MimeType(mime)) {
		return "", fmt.Errorf("invalid content type uploaded: %s", mime)
	}

	ext = mimetypes.GetExtensionFromMime(models.MimeType(mime))
	return ext, nil
}

func stripMimeParams(mime string) string {
	if i := strings.Index(mime, ";"); i != -1 {
		return strings.TrimSpace(mime[:i])
	}
	return mime
}

func (ss *StepService) getPath(base_path string, extension string) (path string, filename string) {
	filename = uuid.Must(uuid.NewV7()).String() + extension
	return filepath.Join(base_path, filename), filename
}

func (ss *StepService) getFlags(task models.Task, step models.Step) (map[string]string, error) {
	flagMap := make(map[string]string)

	for _, flagDef := range task.Flags {

		for _, flagVal := range step.Flags {
			if flagVal.Name == flagDef.Name {
				flagMap[flagVal.Name] = flagVal.Value
				break
			}
		}

		if _, ok := flagMap[flagDef.Name]; !ok && flagDef.Required {
			return flagMap, fmt.Errorf("Flag %s is required", flagDef.Name)
		}

	}

	return flagMap, nil
}
