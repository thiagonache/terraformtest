# Unit testing terraform

![Go](https://github.com/thiagonache/terraformtest/workflows/Go/badge.svg?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/thiagonache/terraformtest)](https://goreportcard.com/report/github.com/thiagonache/terraformtest)

## Import
```go
import "github.com/thiagonache/terraformtest"
```

## Summary
Terraformtest is a lightweight terraform tester written in Go for unit and integration tests.

### Features
- Remove code paperwork to write tests.
- No JSON query required for testing resources.
- Terraform resources are abstract in a Go generic struct.
- Test number of resources wanted in a plan.
- Test if a wanted resource exist in a plan.
- Test if all wanted resources exist in a plan and vice-versa.

## Testing your plan

### Test if plan contains one or more resources

```go
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

    wantNumResources := 40
    items := p.Resources.Resources

    if len(items) != wantNumResources {
        t.Errorf("want %d resources in plan, got %d", wantNumResources, len(items))
    }
```

### Test if plan is equal (have all the resources wanted)

```go
func TestEqual(t *testing.T) {
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
```
