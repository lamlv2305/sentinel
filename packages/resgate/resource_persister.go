package resgate

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ResourceQuery struct {
	ProjectId  string `json:"project_id"`
	Group      string `json:"group,omitempty"`
	ResourceId string `json:"resource_id,omitempty"`
}

type ResourcePersister interface {
	Save(ctx context.Context, resource Resource) error
	Delete(ctx context.Context, query ResourceQuery) error
	Get(ctx context.Context, resourceId string) (Resource, error)
	GetByQuery(ctx context.Context, query ResourceQuery) (Resource, error) // New method
	List(ctx context.Context, query ResourceQuery) ([]Resource, error)
}

var _ ResourcePersister = &FilePersister{}

type FilePersister struct {
	path string
}

func NewFilePersister(path string) ResourcePersister {
	// Ensure the directory exists
	if err := os.MkdirAll(path, 0o755); err != nil {
		panic(fmt.Sprintf("failed to create directory %s: %v", path, err))
	}

	return &FilePersister{
		path: path,
	}
}

// Delete implements ResourcePersister.
func (f *FilePersister) Delete(ctx context.Context, query ResourceQuery) error {
	targetPath := f.buildTargetPath(query)

	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return nil
	}

	if err := os.RemoveAll(targetPath); err != nil {
		return fmt.Errorf("failed to delete %s: %w", targetPath, err)
	}

	f.cleanupEmptyDirs(query)
	return nil
}

func (f *FilePersister) buildTargetPath(query ResourceQuery) string {
	projectDir := filepath.Join(f.path, query.ProjectId)

	if query.Group == "" {
		if query.ResourceId == "" {
			return projectDir
		}
		return filepath.Join(projectDir, query.ResourceId+".json")
	}

	if query.ResourceId == "" {
		return filepath.Join(projectDir, query.Group)
	}
	return filepath.Join(projectDir, query.Group, query.ResourceId+".json")
}

func (f *FilePersister) cleanupEmptyDirs(query ResourceQuery) {
	projectDir := filepath.Join(f.path, query.ProjectId)

	if query.ResourceId != "" && query.Group != "" {
		groupDir := filepath.Join(projectDir, query.Group)
		if f.isDirEmpty(groupDir) {
			os.Remove(groupDir)
		}
	}

	if query.Group != "" || query.ResourceId != "" {
		if f.isDirEmpty(projectDir) {
			os.Remove(projectDir)
		}
	}
}

// Get implements ResourcePersister.
func (f *FilePersister) Get(ctx context.Context, resourceId string) (Resource, error) {
	var found Resource

	err := filepath.Walk(f.path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(info.Name(), ".json") {
			return nil
		}

		if strings.TrimSuffix(info.Name(), ".json") != resourceId {
			return nil
		}

		resource, err := f.loadResourceFromFile(path)
		if err != nil {
			return err
		}

		if resource.ResourceId == resourceId {
			found = resource
			return fmt.Errorf("found")
		}

		return nil
	})

	if err != nil && err.Error() == "found" {
		return found, nil
	}

	if err != nil {
		return Resource{}, fmt.Errorf("error searching for resource: %w", err)
	}

	return Resource{}, fmt.Errorf("resource with id '%s' not found", resourceId)
}

// GetByQuery gets a single resource using a ResourceQuery (more efficient than Get)
func (f *FilePersister) GetByQuery(ctx context.Context, query ResourceQuery) (Resource, error) {
	if query.ProjectId == "" || query.ResourceId == "" {
		return Resource{}, fmt.Errorf("projectId and resourceId are required")
	}

	resourcePath := f.buildResourcePath(query.ProjectId, query.Group, query.ResourceId)
	return f.loadResourceFromFile(resourcePath)
}

func (f *FilePersister) buildResourcePath(projectId, group, resourceId string) string {
	if group == "" {
		return filepath.Join(f.path, projectId, resourceId+".json")
	}
	return filepath.Join(f.path, projectId, group, resourceId+".json")
}

// List implements ResourcePersister.
func (f *FilePersister) List(ctx context.Context, query ResourceQuery) ([]Resource, error) {
	searchPath := f.buildSearchPath(query)

	if _, err := os.Stat(searchPath); os.IsNotExist(err) {
		return []Resource{}, nil
	}

	if query.ResourceId != "" {
		return f.listSingleResource(searchPath, query.ResourceId)
	}

	return f.listAllResources(searchPath)
}

func (f *FilePersister) buildSearchPath(query ResourceQuery) string {
	projectDir := filepath.Join(f.path, query.ProjectId)
	if query.Group == "" {
		return projectDir
	}
	return filepath.Join(projectDir, query.Group)
}

func (f *FilePersister) listSingleResource(searchPath, resourceId string) ([]Resource, error) {
	resourcePath := filepath.Join(searchPath, resourceId+".json")
	resource, err := f.loadResourceFromFile(resourcePath)
	if os.IsNotExist(err) {
		return []Resource{}, nil
	}
	if err != nil {
		return nil, err
	}
	return []Resource{resource}, nil
}

func (f *FilePersister) listAllResources(searchPath string) ([]Resource, error) {
	var resources []Resource

	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(info.Name(), ".json") {
			return nil
		}

		resource, err := f.loadResourceFromFile(path)
		if err != nil {
			return fmt.Errorf("failed to load resource from %s: %w", path, err)
		}

		resources = append(resources, resource)
		return nil
	})

	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return resources, nil
}

// Save implements ResourcePersister.
func (f *FilePersister) Save(ctx context.Context, resource Resource) error {
	resourcePath := f.buildResourcePath(resource.ProjectId, resource.Group, resource.ResourceId)

	if err := os.MkdirAll(filepath.Dir(resourcePath), 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := json.MarshalIndent(resource, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal resource: %w", err)
	}

	if err := os.WriteFile(resourcePath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write resource file: %w", err)
	}

	return nil
}

// loadResourceFromFile loads a resource from a JSON file
func (f *FilePersister) loadResourceFromFile(filePath string) (Resource, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return Resource{}, err
	}

	var resource Resource
	if err := json.Unmarshal(data, &resource); err != nil {
		return Resource{}, fmt.Errorf("failed to unmarshal resource from %s: %w", filePath, err)
	}

	return resource, nil
}

// isDirEmpty checks if a directory is empty
func (f *FilePersister) isDirEmpty(dirPath string) bool {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return false
	}
	return len(entries) == 0
}
