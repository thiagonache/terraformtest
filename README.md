![Go](https://github.com/thiagonache/terraformtest/workflows/Go/badge.svg?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/thiagonache/terraformtest)](https://goreportcard.com/report/github.com/thiagonache/terraformtest)


# Unit testing terraform (WIP)

## Disclaimer

Currently, the only way to compare values is using JSON query path and all types
are strings.

```
  want := terraformtest.TFResource{
    Filter: `planned_values.root_module.child_modules.#.resources`,
    Check: map[string]string{
      "0.0.address":              "module.nomad_job.nomad_job.test_job",
      "0.0.type":                 "nomad_job",
      "0.0.values.name":          "unit-test",
      "0.0.values.datacenters.0": "dc1",
    },
  }
```

If we would have a second datacenter we would add one more line.

```
  want := terraformtest.TFResource{
    Filter: `planned_values.root_module.child_modules.#.resources`,
    Check: map[string]string{
      "0.0.address":              "module.nomad_job.nomad_job.test_job",
      "0.0.type":                 "nomad_job",
      "0.0.values.name":          "unit-test",
      "0.0.values.datacenters.0": "dc1",
      "0.0.values.datacenters.1": "dc2",
    },
  }
```
