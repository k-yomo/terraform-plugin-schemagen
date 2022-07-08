package terraform_plugin_schemagen

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	install "github.com/hashicorp/hc-install"
	"github.com/hashicorp/hc-install/checkpoint"
	"github.com/hashicorp/hc-install/fs"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/src"
	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
)

func TerraformProviderSchema(ctx context.Context) (*tfjson.ProviderSchema, error) {
	tmpDir, err := ioutil.TempDir("", "tfws")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	shortName := "schemagen-temp"
	providerPath := fmt.Sprintf("plugins/registry.terraform.io/hashicorp/%s/0.0.1/%s_%s", shortName, runtime.GOOS, runtime.GOARCH)
	outFile := filepath.Join(tmpDir, providerPath, fmt.Sprintf("terraform-provider-%s", shortName))
	switch runtime.GOOS {
	case "windows":
		outFile = outFile + ".exe"
	}
	buildCmd := exec.Command("go", "build", "-o", outFile)
	_, err = runCmd(buildCmd)
	if err != nil {
		return nil, fmt.Errorf("execute build command: %w", err)
	}

	err = writeFile(filepath.Join(tmpDir, "provider.tf"), fmt.Sprintf(`
provider %[1]q {
}
`, shortName))
	if err != nil {
		return nil, fmt.Errorf("write temp provider: %w", err)
	}

	i := install.NewInstaller()
	sources := []src.Source{
		&fs.AnyVersion{
			Product: &product.Terraform,
		},
		&checkpoint.LatestVersion{
			InstallDir: tmpDir,
			Product:    product.Terraform,
		},
	}

	tfBin, err := i.Ensure(ctx, sources)
	if err != nil {
		return nil, fmt.Errorf("ensure source: %w", err)
	}

	tf, err := tfexec.NewTerraform(tmpDir, tfBin)
	if err != nil {
		return nil, err
	}

	err = tf.Init(ctx, tfexec.Get(false), tfexec.PluginDir("./plugins"))
	if err != nil {
		return nil, err
	}

	schemas, err := tf.ProvidersSchema(ctx)
	if err != nil {
		return nil, err
	}

	if ps, ok := schemas.Schemas[shortName]; ok {
		return ps, nil
	}

	if ps, ok := schemas.Schemas["registry.terraform.io/hashicorp/"+shortName]; ok {
		return ps, nil
	}

	return nil, fmt.Errorf("unable to find schema in JSON for provider %q", shortName)
}

func runCmd(cmd *exec.Cmd) ([]byte, error) {
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("error executing %q, %v", cmd.Path, cmd.Args)
		log.Printf(string(output))
		return nil, fmt.Errorf("error executing %q: %w", cmd.Path, err)
	}
	return output, nil
}

func writeFile(path string, data string) error {
	dir, _ := filepath.Split(path)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("unable to make dir %q: %w", dir, err)
	}

	err = ioutil.WriteFile(path, []byte(data), 0644)
	if err != nil {
		return fmt.Errorf("unable to write file %q: %w", path, err)
	}

	return nil
}
