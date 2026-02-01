package project

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type ControllerDef struct {
	Name string
	URL  string
}

func PushProjectToController(name string) error {
	ctrl, err := getControllerFromName(name)
	if err != nil {
		return err
	}

	proj, err := GetProject()
	if err != nil {
		return err
	}

	request_url, err := url.JoinPath(ctrl.URL, "api", "v1", "project", proj.ID.String())
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(proj)
	if err != nil {
		return err
	}

	s, err := http.Post(request_url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer s.Body.Close()

	if s.StatusCode < 200 || s.StatusCode > 299 {
		errorMessage, _ := io.ReadAll(s.Body)
		return fmt.Errorf("error pushing project: %s", string(errorMessage))
	}

	return err
}

func AddController(name, url string) error {
	root, err := GetProjectRoot()
	if err != nil {
		return err
	}

	proj, err := getProjectFile(root)
	if err != nil {
		return err
	}

	for _, ctrl := range proj.Controllers {
		if ctrl.Name == name {
			return fmt.Errorf("controller with name %s already exists", name)
		}
	}

	proj.Controllers = append(proj.Controllers, ControllerDef{
		Name: name,
		URL:  url,
	})

	return writeProjectFile(root, *proj)
}

func ListControllers() ([]ControllerDef, error) {
	proj, err := GetProjectFile()
	if err != nil {
		return nil, err
	}

	return proj.Controllers, nil
}

func getControllerFromName(name string) (ControllerDef, error) {
	proj, err := GetProjectFile()
	if err != nil {
		return ControllerDef{}, err
	}

	for _, ctrl := range proj.Controllers {
		if ctrl.Name == name {
			return ctrl, nil
		}
	}

	return ControllerDef{}, fmt.Errorf("controller with name %s not found", name)
}
