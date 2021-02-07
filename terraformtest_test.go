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

	want := 10

	tfPlan, err := terraformtest.ParseJSON("terraform.tfplan")
	if err != nil {
		t.Error(err)
	}
	got, err := terraformtest.CountNumberResources(tfPlan, ".planned_values.root_module.child_modules.[0]")
	if err != nil {
		t.Fatal(err)
	}

	if want != got {
		t.Errorf("want %d, got %d", want, got)
	}
}
