package project

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/pupload/pupload/internal/models"

	"sigs.k8s.io/yaml"
)

type ProjectFile struct {
	ID          uuid.UUID `yaml:"id"`
	ProjectName string    `yaml:"project_name"`

	Controllers  []ControllerDef     `yaml:"Controllers"`
	GlobalStores []models.StoreInput `yaml:"GlobalStores"`

	Extra map[string]any `yaml:",inline"`
}

func newProjectFile(projectName string) ProjectFile {
	return ProjectFile{
		ID:          uuid.Must(uuid.NewV7()),
		ProjectName: projectName,
		Controllers: []ControllerDef{},

		GlobalStores: []models.StoreInput{},
	}
}

func TestFlow(projectRoot, controllerAddress, flowName string) (*models.FlowRun, *models.Flow, error) {

	flow, err := GetFlow(projectRoot, flowName)
	if err != nil {
		return nil, nil, err
	}

	flow.Normalize()

	node_defs, err := GetNodeDefs(projectRoot)
	if err != nil {
		return nil, nil, err
	}

	body := struct {
		Flow     models.Flow      `json:"Flow"`
		NodeDefs []models.NodeDef `json:"NodeDefs"`
	}{
		Flow:     *flow,
		NodeDefs: node_defs,
	}

	j, err := json.Marshal(&body)
	if err != nil {
		return nil, nil, err
	}

	url, _ := url.JoinPath(controllerAddress, "api", "v1", "flow", "test")

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(j))
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)

		return nil, nil, fmt.Errorf("TestFlow: Controller could not run flow: %s", string(body))
	}

	flow_run := new(models.FlowRun)
	err = json.NewDecoder(resp.Body).Decode(flow_run)

	return flow_run, flow, nil
}

func GetFlow(projectRoot, flowName string) (*models.Flow, error) {

	flows, _ := GetFlows(projectRoot)

	var flow *models.Flow

	for _, f := range flows {
		if f.Name == flowName {
			flow = &f
			break
		}
	}

	if flow == nil {
		return nil, fmt.Errorf("flow %s not found", flowName)
	}

	return flow, nil
}

func GetNodeDefs(projectRoot string) ([]models.NodeDef, error) {
	path := filepath.Join(projectRoot, "node_defs")
	nodeDefs := make([]models.NodeDef, 0)

	yamls, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, y := range yamls {
		var nodeDef models.NodeDef
		data, err := os.ReadFile(filepath.Join(path, y.Name()))
		if err != nil {
			continue
		}

		if err := yaml.Unmarshal(data, &nodeDef); err != nil {
			continue
		}

		nodeDefs = append(nodeDefs, nodeDef)
	}

	return nodeDefs, nil
}

func loadFlowsFromDir(path string) ([]models.Flow, error) {
	flows := make([]models.Flow, 0)

	yamls, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, y := range yamls {
		var flow models.Flow
		data, err := os.ReadFile(filepath.Join(path, y.Name()))
		if err != nil {
			continue
		}

		if err := yaml.Unmarshal(data, &flow); err != nil {
			continue
		}

		flows = append(flows, flow)
	}

	return flows, nil
}

func loadDefsFromDir(path string) ([]models.NodeDef, error) {
	nodeDefs := make([]models.NodeDef, 0)

	yamls, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, y := range yamls {
		var nodeDef models.NodeDef
		data, err := os.ReadFile(filepath.Join(path, y.Name()))
		if err != nil {
			continue
		}

		if err := yaml.Unmarshal(data, &nodeDef); err != nil {
			continue
		}

		nodeDefs = append(nodeDefs, nodeDef)
	}

	return nodeDefs, nil
}

func GetFlows(projectRoot string) ([]models.Flow, error) {
	path := filepath.Join(projectRoot, "flows")
	flows := make([]models.Flow, 0)

	yamls, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, y := range yamls {
		var flow models.Flow
		data, err := os.ReadFile(filepath.Join(path, y.Name()))
		if err != nil {
			continue
		}

		if err := yaml.Unmarshal(data, &flow); err != nil {
			continue
		}

		flows = append(flows, flow)
	}

	return flows, nil
}

func GetProjectFile() (*ProjectFile, error) {
	root, err := GetProjectRoot()
	if err != nil {
		return nil, err
	}

	return getProjectFile(root)

}

func getProjectFile(projectRoot string) (*ProjectFile, error) {

	projectPath := filepath.Join(projectRoot, "pup.yaml")
	file, err := os.ReadFile(projectPath)
	if err != nil {
		return nil, err
	}

	var project ProjectFile
	if err := yaml.Unmarshal(file, &project); err != nil {
		return nil, err
	}

	return &project, nil
}

func InitProject(path string, projectName string) error {

	project := newProjectFile(projectName)
	if err := writeProjectFile(path, project); err != nil {
		return err
	}

	if err := os.Mkdir("flows", 0600); err != nil {
		return err
	}

	if err := os.Mkdir("node_defs", 0600); err != nil {
		return err
	}

	return nil
}

func writeProjectFile(projectRoot string, projectFile ProjectFile) error {

	file := filepath.Join(projectRoot, "pup.yaml")

	yamlBytes, err := yaml.Marshal(&projectFile)
	if err != nil {
		return err
	}

	return os.WriteFile(file, yamlBytes, 0600)

}

func isProjectDirectory(path string) bool {
	if _, err := ProjectRoot(path); err != nil {
		return false
	}

	return true
}

func GetProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	root, err := ProjectRoot(dir)
	if err != nil {
		return "", err
	}

	return root, nil

}

func ProjectRoot(path string) (string, error) {

	for {
		entries, err := os.ReadDir(path)
		if err != nil {
			return "", err
		}

		for _, e := range entries {
			if e.Name() == "pup.yaml" {
				return filepath.Abs(path)
			}
		}

		path = filepath.Join(path, "..")
	}
}

func GetProject() (models.Project, error) {
	root, err := GetProjectRoot()
	if err != nil {
		return models.Project{}, err
	}

	return getProject(root)
}

func getProject(projectRoot string) (models.Project, error) {
	flows, err := GetFlows(projectRoot)
	if err != nil {
		return models.Project{}, err
	}

	nodeDefs, err := GetNodeDefs(projectRoot)
	if err != nil {
		return models.Project{}, err
	}

	projectFile, err := getProjectFile(projectRoot)
	if err != nil {
		return models.Project{}, err
	}

	return models.Project{
		ID:       projectFile.ID,
		Flows:    flows,
		NodeDefs: nodeDefs,
	}, nil
}
