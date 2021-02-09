package terraformtest_test

import (
	"terraformtest"
	"testing"
)

func TestJSONParse(t *testing.T) {
	t.Parallel()

	wantBiggerThan := 1000
	got, err := terraformtest.ParseJSON("terraform.tfplan")
	if err != nil {
		t.Error(err)
	}
	if wantBiggerThan >= len(got) {
		t.Errorf("want json bigger than %d but got %d", wantBiggerThan, len(got))
	}
}

func TestNumberResources(t *testing.T) {
	t.Parallel()

	want := 2
	tfPlan, err := terraformtest.ParseJSON("terraform.tfplan2")
	if err != nil {
		t.Error(err)
	}
	got, err := terraformtest.CountNumberResources(tfPlan, ".planned_values.root_module.resources")
	if err != nil {
		t.Fatal(err)
	}

	if want != got {
		t.Errorf("want %d, got %d", want, got)
	}
}

func TestCompareNomadJob(t *testing.T) {
	t.Parallel()
	want := map[string]string{"name": "test_job"}

	tfPlan, err := terraformtest.ParseJSON("terraform.tfplan")
	if err != nil {
		t.Error(err)
	}
	got, err := terraformtest.ExtractPlanData(tfPlan, ".planned_values.root_module.child_modules.[0].resources.[0]")
	if !terraformtest.Equal(want, got) {
		t.Fatal("not equal")
	}
}
