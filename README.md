# terraform-plugin-schemagen
Generate Go helper object for Terraform provider schema.


## Usage
1. Add the following line to `main.go` in a terraform provider repository.
```diff
//go:generate go run github.com/k-yomo/terraform-plugin-schemagen/cmd/tfschemagen
```

2. `go generate ./...` will generate schema code in `schemagen/schema_gen.go`.

3. Use schema helper code for a resource.
```go
func resourceTestCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
    testData := schemagen.NewTest(d)
    requiredField := testData.MustRequiredField()
    optionalField, ok := testData.OptionalField()
    // make API call with the fields
}
```
