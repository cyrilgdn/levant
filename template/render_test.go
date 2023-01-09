package template

import (
	"os"
	"testing"

	"github.com/hashicorp/levant/levant/structs"
	nomad "github.com/hashicorp/nomad/api"
)

const (
	testJobName           = "levantExample"
	testJobNameOverwrite  = "levantExampleOverwrite"
	testJobNameOverwrite2 = "levantExampleOverwrite2"
	testDCName            = "dc13"
	testEnvName           = "GROUP_NAME_ENV"
	testEnvValue          = "cache"
)

func TestTemplater_RenderTemplate(t *testing.T) {

	var job *nomad.Job
	var err error

	// Start with an empty passed var args map.
	fVars := make(map[string]interface{})

	// Test basic TF template render.
	config := &structs.TemplateConfig{
		TemplateFile:  "test-fixtures/single_templated.nomad",
		VariableFiles: []string{"test-fixtures/test.tf"},
		DisableHCL2:   true,
	}
	job, err = RenderJob(config, "", &fVars)
	if err != nil {
		t.Fatal(err)
	}
	if *job.Name != testJobName {
		t.Fatalf("expected %s but got %v", testJobName, *job.Name)
	}
	if *job.TaskGroups[0].Tasks[0].Resources.CPU != 1313 {
		t.Fatalf("expected CPU resource %v but got %v", 1313, *job.TaskGroups[0].Tasks[0].Resources.CPU)
	}

	// Test basic YAML template render.
	config = &structs.TemplateConfig{TemplateFile: "test-fixtures/single_templated.nomad", VariableFiles: []string{"test-fixtures/test.yaml"}}
	job, err = RenderJob(config, "", &fVars)
	if err != nil {
		t.Fatal(err)
	}
	if *job.Name != testJobName {
		t.Fatalf("expected %s but got %v", testJobName, *job.Name)
	}
	if *job.TaskGroups[0].Tasks[0].Resources.CPU != 1313 {
		t.Fatalf("expected CPU resource %v but got %v", 1313, *job.TaskGroups[0].Tasks[0].Resources.CPU)
	}

	// Test multiple var-files
	config = &structs.TemplateConfig{
		TemplateFile:  "test-fixtures/single_templated.nomad",
		VariableFiles: []string{"test-fixtures/test.yaml", "test-fixtures/test-overwrite.yaml"},
	}
	job, err = RenderJob(config, "", &fVars)
	if err != nil {
		t.Fatal(err)
	}
	if *job.Name != testJobNameOverwrite {
		t.Fatalf("expected %s but got %v", testJobNameOverwrite, *job.Name)
	}

	// Test multiple var-files of different types
	config = &structs.TemplateConfig{
		TemplateFile:  "test-fixtures/single_templated.nomad",
		VariableFiles: []string{"test-fixtures/test.tf", "test-fixtures/test-overwrite.yaml"},
	}
	job, err = RenderJob(config, "", &fVars)
	if err != nil {
		t.Fatal(err)
	}
	if *job.Name != testJobNameOverwrite {
		t.Fatalf("expected %s but got %v", testJobNameOverwrite, *job.Name)
	}

	// Test multiple var-files with var-args
	fVars["job_name"] = testJobNameOverwrite2

	config = &structs.TemplateConfig{
		TemplateFile:  "test-fixtures/single_templated.nomad",
		VariableFiles: []string{"test-fixtures/test.tf", "test-fixtures/test-overwrite.yaml"},
	}
	job, err = RenderJob(config, "", &fVars)
	if err != nil {
		t.Fatal(err)
	}
	if *job.Name != testJobNameOverwrite2 {
		t.Fatalf("expected %s but got %v", testJobNameOverwrite2, *job.Name)
	}

	// Test empty var-args and empty variable file render.
	config = &structs.TemplateConfig{TemplateFile: "test-fixtures/none_templated.nomad", VariableFiles: []string{}}
	job, err = RenderJob(config, "", &fVars)
	if err != nil {
		t.Fatal(err)
	}
	if *job.Name != testJobName {
		t.Fatalf("expected %s but got %v", testJobName, *job.Name)
	}

	// Test var-args only render.
	config = &structs.TemplateConfig{TemplateFile: "test-fixtures/single_templated.nomad", VariableFiles: []string{}}
	fVars = map[string]interface{}{"job_name": testJobName, "task_resource_cpu": "1313"}
	job, err = RenderJob(config, "", &fVars)
	if err != nil {
		t.Fatal(err)
	}
	if *job.Name != testJobName {
		t.Fatalf("expected %s but got %v", testJobName, *job.Name)
	}
	if *job.TaskGroups[0].Tasks[0].Resources.CPU != 1313 {
		t.Fatalf("expected CPU resource %v but got %v", 1313, *job.TaskGroups[0].Tasks[0].Resources.CPU)
	}

	// Test var-args only render with HCL2 spec
	config = &structs.TemplateConfig{TemplateFile: "test-fixtures/single_templated_connect.nomad"}
	fVars = map[string]interface{}{"job_name": testJobName, "task_resource_cpu": "1313", "upstream_datacenter": "dc2"}
	job, err = RenderJob(config, "", &fVars)
	if err != nil {
		t.Fatal(err)
	}
	if *job.Name != testJobName {
		t.Fatalf("expected %s but got %v", testJobName, *job.Name)
	}
	if *job.TaskGroups[0].Tasks[0].Resources.CPU != 1313 {
		t.Fatalf("expected CPU resource %v but got %v", 1313, *job.TaskGroups[0].Tasks[0].Resources.CPU)
	}
	if job.TaskGroups[0].Services[0].Connect.SidecarService.Proxy.Upstreams[0].Datacenter != "dc2" {
		t.Fatalf("expected connect upstream datacenter %v but got %v", "dc2", job.TaskGroups[0].Services[0].Connect.SidecarService.Proxy.Upstreams[0].Datacenter)
	}

	// Test var-args and variables file render.
	delete(fVars, "job_name")
	config = &structs.TemplateConfig{TemplateFile: "test-fixtures/multi_templated.nomad", VariableFiles: []string{"test-fixtures/test.yaml"}}
	fVars["datacentre"] = testDCName
	os.Setenv(testEnvName, testEnvValue)
	job, err = RenderJob(config, "", &fVars)
	if err != nil {
		t.Fatal(err)
	}
	if *job.Name != testJobName {
		t.Fatalf("expected %s but got %v", testJobName, *job.Name)
	}
	if job.Datacenters[0] != testDCName {
		t.Fatalf("expected %s but got %v", testDCName, job.Datacenters[0])
	}
	if *job.TaskGroups[0].Name != testEnvValue {
		t.Fatalf("expected %s but got %v", testEnvValue, *job.TaskGroups[0].Name)
	}
}
