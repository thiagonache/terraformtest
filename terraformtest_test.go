package terraformtest_test

import (
	"terraformtest"
	"testing"
)

func TestReadPlanFile(t *testing.T) {
	t.Parallel()

	wantLen := 8716
	p, err := terraformtest.ReadPlan("testdata/terraform.plan.json")
	if err != nil {
		t.Fatal(err)
	}

	if wantLen != len(p.PlanData) {
		t.Errorf("want json size in bytes of %d but got %d", wantLen, len(p.PlanData))
	}
}

func TestNumberResources(t *testing.T) {
	t.Parallel()

	wantNumResources := 40

	p, err := terraformtest.ReadPlan("testdata/terraform-aws-101.plan.json")
	if err != nil {
		t.Fatal(err)
	}
	items := p.Resources.Items

	if len(items) != wantNumResources {
		t.Errorf("want %d resources in plan, got %d", wantNumResources, len(items))
	}
}

func TestEqual(t *testing.T) {
	t.Parallel()

	wantRS := []terraformtest.Resource{
		{
			Address: "module.nomad_job.nomad_job.test_job",
			Metadata: map[string]string{
				"type": "nomad_job",
				"name": "test_job",
			},
			Values: map[string]string{
				"name":        "unit-test",
				"datacenters": `["dc1"]`,
			},
		},
	}

	p, err := terraformtest.ReadPlan("testdata/terraform.plan.json")
	if err != nil {
		t.Fatal(err)
	}
	gotRS := p.Resources

	if !terraformtest.Equal(wantRS, &gotRS) {
		t.Error(gotRS.Diff())
	}
}

func TestContains(t *testing.T) {
	t.Parallel()

	wantRS := terraformtest.Resource{
		Address: "module.vpc.aws_eip.nat[0]",
		Metadata: map[string]string{
			"type": "aws_eip",
			"name": "nat",
		},
		Values: map[string]string{
			"vpc":      "true",
			"timeouts": "",
		},
	}

	p, err := terraformtest.ReadPlan("testdata/terraform-aws-101.plan.json")
	if err != nil {
		t.Fatal(err)
	}
	gotRS := p.Resources

	if !gotRS.Contains(wantRS) {
		t.Error(gotRS.Diff())
	}
}

func TestContainsResource(t *testing.T) {
	testCases := []struct {
		desc, planJSONPath string
		wantResource       terraformtest.Resource
	}{
		{
			desc:         "Test EIP",
			planJSONPath: "testdata/terraform-aws-101.plan.json",
			wantResource: terraformtest.Resource{
				Address: "module.vpc.aws_eip.nat[0]",
				Metadata: map[string]string{
					"type": "aws_eip",
					"name": "nat",
				},
				Values: map[string]string{
					"vpc":      "true",
					"timeouts": "",
				},
			},
		},
		{
			desc:         "Test DB Subnet Group",
			planJSONPath: "testdata/terraform-aws-101.plan.json",
			wantResource: terraformtest.Resource{
				Address: "module.db.module.db_subnet_group.aws_db_subnet_group.this[0]",
				Metadata: map[string]string{
					"type": "aws_db_subnet_group",
					"name": "this",
				},
				Values: map[string]string{
					"name_prefix": "demodb-",
				},
			},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			p, err := terraformtest.ReadPlan(tC.planJSONPath)
			if err != nil {
				t.Fatal(err)
			}
			gotRS := p.Resources
			if !gotRS.Contains(tC.wantResource) {
				t.Error(gotRS.Diff())
			}
		})
	}
}
