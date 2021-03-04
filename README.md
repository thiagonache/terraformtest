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

### Motivation

![Martin's Folwer Pyramid
Tests](https://3fxtqy18kygf3on3bu39kh93-wpengine.netdna-ssl.com/wp-content/uploads/2020/01/test-automation-pyramid.jpg)

- The most unit tests the better so they must be fast.

- Should be simple to write tests so the writer doesn't need to be an expert on Go
  or Terraform.

## Testing your plan

### Generate Terraform plan in JSON format

1. Run plan with plan in binary

   ```shell
   terraform plan -out /tmp/mymodule.plan
   ```

1. Convert binary plan into JSON file.

   ```shell
   terraform show -json /tmp/mymodule.plan > /any/path/i/want/mymodule.plan.json
   ```

### Test if plan contains one or more resources

```go
func TestContainsResource(t *testing.T) {
    testCases := []struct {
        desc, planJSONPath string
        wantResource       terraformtest.Resource
    }{
        {
            // Test description
            desc:         "Test EIP",
            // Path for Terraform plan with resources to be tested in JSON format.
            planJSONPath: "testdata/terraform-aws-101.plan.json",
            wantResource: terraformtest.Resource{
                // Resource address as show in the output of terraform plan command.
                Address: "module.vpc.aws_eip.nat[0]",
                // Metadata represents the resource type and the resource name in the resource declaration.
                // Eg.: resource "aws_eip" "nat" {
                Metadata: map[string]string{
                    "type": "aws_eip",
                    "name": "nat",
                },
                // Values are the resources key => value. Anything inside of the planned_values in the JSON file.
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
            // Read the Terraform JSON plan file
            p, err := terraformtest.ReadPlan(tC.planJSONPath)
            if err != nil {
                t.Fatal(err)
            }
            // Get the resourceSet (map of resources in the plan)
            gotRS := p.Resources
            // Test if the resource wanted is present in the plan
            if !gotRS.Contains(tC.wantResource) {
                // If does not contain, output the diff
                t.Error(gotRS.Diff())
            }
        })
    }

    // Set the total number of resources the plan must have
    wantNumResources := 40
    items := p.Resources.Resources

    // Test if number of resources in the plan is equal to number of resources wanted
    if len(items) != wantNumResources {
        t.Errorf("want %d resources in plan, got %d", wantNumResources, len(items))
    }
}
```

### Test if plan is equal (have all the resources wanted)

The difference between Contains and Equal is that Equal requires all resources
to be declared in the slice of wanted Resource. If there's one item in the plan
that doesn't exist in the resourceSet or vice-versa it fails.

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
}
```
