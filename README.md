![Go](https://github.com/thiagonache/terraformtest/workflows/Go/badge.svg?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/thiagonache/terraformtest)](https://goreportcard.com/report/github.com/thiagonache/terraformtest)


# Unit testing terraform (WIP)

## Disclaimer

Currently, the only way to compare values is using JSON query path and all types
are strings.

```
  want := terraformtest.TFResource{
    Metadata: map[string]string{
      "name": "test_job",
      "type": "nomad_job",
    },
    Total: 1,
    Values: map[string]string{
      "name":          "unit-test",
      "datacenters.0": "dc1",
    },
  }
```

If we would have a second datacenter we would add one more line inside of Values
field.

```
    Values: map[string]string{
      "name":          "unit-test",
      "datacenters.0": "dc1",
      "datacenters.1": "dc2",
    },
```
