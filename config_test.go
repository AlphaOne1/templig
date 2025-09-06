// Copyright the templig contributors.
// SPDX-License-Identifier: MPL-2.0

package templig_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/AlphaOne1/templig"
)

type TestConn struct {
	URL    string   `yaml:"url"`
	Passes []string `yaml:"passes"`
}

type TestConfig struct {
	ID   int       `yaml:"id"`
	Name string    `yaml:"name"`
	Conn *TestConn `yaml:"conn,omitempty"`
}

func TestReadConfig(t *testing.T) {
	tests := []struct {
		in      string
		inFile  string
		env     map[string]string
		args    []string
		want    TestConfig
		wantErr bool
	}{
		{ // 0
			inFile:  "testData/test_empty.yaml",
			want:    TestConfig{},
			wantErr: true,
		},
		{ // 1
			in: `
                name: "Name0"`,
			want: TestConfig{
				Name: "Name0",
			},
			wantErr: false,
		},
		{ // 2
			in: `
                id: 23`,
			want: TestConfig{
				ID: 23,
			},
			wantErr: false,
		},
		{ // 3
			in: `
                id:   23
                name: Name0`,
			want: TestConfig{
				ID:   23,
				Name: "Name0",
			},
			wantErr: false,
		},
		{ // 4
			in: `
                id:   23
                name: Name0
                conn:
                  url: https://www.tests.to
                  passes:
                    - password0
                    - password1`,
			want: TestConfig{
				ID:   23,
				Name: "Name0",
				Conn: &TestConn{
					URL: "https://www.tests.to",
					Passes: []string{
						"password0",
						"password1",
					},
				},
			},
			wantErr: false,
		},
		{ // 5
			in: `
                id:   23
                name: {{ required "has to be set" "Name0" | quote }}
                conn:
                  url: https://www.tests.to
                  passes:
                    - password0
                    - password1`,
			want: TestConfig{
				ID:   23,
				Name: "Name0",
				Conn: &TestConn{
					URL: "https://www.tests.to",
					Passes: []string{
						"password0",
						"password1",
					},
				},
			},
			wantErr: false,
		},
		{ // 6
			in: `
                id:   23
                name: {{ required "has to be set" "" | quote }}`,
			want: TestConfig{
				ID:   23,
				Name: "Name0",
			},
			wantErr: true,
		},
		{ // 7
			in: `
                id:   23
                name: {{ required "has to be set" nil | quote }}`,
			want: TestConfig{
				ID:   23,
				Name: "Name0",
			},
			wantErr: true,
		},
		{ // 8
			in: `
                id:   23
                name: {{ required "has to be set" 9 | quote }}`,
			want: TestConfig{
				ID:   23,
				Name: "9",
			},
			wantErr: false,
		},
		{ // 9
			in: `
                id:   23
                name: {{ required "has to be set" 9 | quote`,
			want: TestConfig{
				ID:   23,
				Name: "9",
			},
			wantErr: true,
		},
		{ // 10
			inFile: "testData/test_config_0.yaml",
			want: TestConfig{
				ID:   9,
				Name: "Name0",
				Conn: &TestConn{
					URL: "https://www.tests.to",
					Passes: []string{
						"pass0",
						"pass1",
					},
				},
			},
			wantErr: false,
		},
		{ // 11
			inFile: "testData/test_config_1.yaml",
			env: map[string]string{
				"PASS1": "pass1",
			},
			want: TestConfig{
				ID:   9,
				Name: "Name1",
				Conn: &TestConn{
					URL: "https://www.tests.to",
					Passes: []string{
						"pass0",
						"pass1",
					},
				},
			},
			wantErr: false,
		},
		{ // 12
			inFile: "testData/test_config_2.yaml",
			want: TestConfig{
				ID:   9,
				Name: "Name1",
				Conn: &TestConn{
					URL: "https://www.tests.to",
					Passes: []string{
						"pass0",
						"cannot_work",
					},
				},
			},
			wantErr: true,
		},
		{ // 13
			in: `
                id:   23
                name: {{ arg "param0" | quote }}`,
			want: TestConfig{
				ID:   23,
				Name: "paramVal0",
			},
			args:    []string{"-param0", "paramVal0"},
			wantErr: false,
		},
		{ // 14
			in: `
                id:   23
                name: {{ arg "param0" | quote }}`,
			want: TestConfig{
				ID:   23,
				Name: "paramVal0",
			},
			args:    []string{"--param0", "paramVal0"},
			wantErr: false,
		},
		{ // 15
			in: `
                id:   23
                name: {{ arg "param0" | quote }}`,
			want: TestConfig{
				ID:   23,
				Name: "paramVal0",
			},
			args:    []string{"--param0=paramVal0"},
			wantErr: false,
		},
		{ // 16
			in: `
                id:   23
                name: {{ arg "param0" | required "param0 required" | quote }}`,
			want: TestConfig{
				ID:   23,
				Name: "true",
			},
			args:    []string{"--param0"},
			wantErr: true,
		},
		{ // 17
			in: `
                id:   23
                name: {{ if hasArg "param0" }} "have" {{ else }} "have not" {{ end }}`,
			want: TestConfig{
				ID:   23,
				Name: "have",
			},
			args:    []string{"--param0"},
			wantErr: false,
		},
		{ // 18
			in: `
                id:   23
                name: {{ if hasArg "param1" }} "have" {{ else }} "have not" {{ end }}`,
			want: TestConfig{
				ID:   23,
				Name: "have not",
			},
			args:    []string{"--param0"},
			wantErr: false,
		},
		{ // 19
			in: `
                id:   23
                name: {{ if hasArg "param0" }} "have" {{ else }} "have not" {{ end }}`,
			want: TestConfig{
				ID:   23,
				Name: "have not",
			},
			args:    []string{"--param00"},
			wantErr: false,
		},
	}

	for testIndex, test := range tests {
		t.Run(fmt.Sprintf("TestReadConfig-%d", testIndex), func(t *testing.T) {
			testBuf := bytes.Buffer{}

			if len(test.in) > 0 && len(test.inFile) > 0 {
				t.Errorf("%v: input data and file given at the same time", testIndex)
			}

			testBuf.WriteString(test.in)

			if test.env != nil {
				for ei, ev := range test.env {
					t.Setenv(ei, ev)
				}
			}

			if test.args != nil {
				os.Args = append(os.Args, test.args...)
			}

			var config *templig.Config[TestConfig]
			var fromErr error

			switch {
			case len(test.in) > 0:
				config, fromErr = templig.From[TestConfig](&testBuf)
			case len(test.inFile) > 0:
				config, fromErr = templig.FromFile[TestConfig](test.inFile)
			default:
				t.Errorf("%v: neither input data nor input file given", testIndex)
			}

			if test.wantErr && fromErr == nil {
				t.Errorf("%v: wanted error but got nil", testIndex)
			}
			if !test.wantErr && fromErr != nil {
				t.Errorf("%v: did not want error but got %v", testIndex, fromErr)
			}

			if config != nil {
				if config.Get().ID != test.want.ID {
					t.Errorf("%v: wanted ID %v but got %v", testIndex, test.want.ID, config.Get().ID)
				}
				if config.Get().Name != test.want.Name {
					t.Errorf("%v: wanted Name %v but got %v", testIndex, test.want.Name, config.Get().Name)
				}
				if (config.Get().Conn != nil) != (test.want.Conn != nil) {
					t.Errorf("%v: wanted Conn == nil -> %v but got %v", testIndex,
						test.want.Conn != nil,
						config.Get().Conn != nil)
				}
				if config.Get().Conn != nil && test.want.Conn != nil {
					if config.Get().Conn.URL != test.want.Conn.URL {
						t.Errorf("%v: wanted URL %v but got %v", testIndex, test.want.Conn.URL, config.Get().Conn.URL)
					}
					for _, p := range test.want.Conn.Passes {
						if !slices.Contains(config.Get().Conn.Passes, p) {
							t.Errorf("%v: wanted passes to containt %v but was not there", testIndex, p)
						}
					}
					for _, p := range config.Get().Conn.Passes {
						if !slices.Contains(test.want.Conn.Passes, p) {
							t.Errorf("%v: found pass %v but should not there", testIndex, p)
						}
					}
				}
			}

			if test.env != nil {
				for ei := range test.env {
					_ = os.Unsetenv(ei)
				}
			}

			if len(test.args) > 0 {
				os.Args = os.Args[:len(os.Args)-len(test.args)]
			}
		})
	}
}

func TestNoReaders(t *testing.T) {
	c, fromErr := templig.From[TestConfig]()

	if fromErr == nil {
		t.Errorf("reading from broken reader should have returned an error")
	}

	if c != nil {
		t.Errorf("reading from broken reader should have returned nil")
	}
}

func TestReadOverlayConfig(t *testing.T) {
	config, configErr := templig.FromFile[TestConfig](
		"testData/test_config_0.yaml",
		"testData/test_config_0_overlay.yaml",
	)

	if configErr != nil {
		t.Errorf("no error expected reading multiple files: %v", configErr)
	}

	if len(config.Get().Conn.Passes) != 3 {
		t.Errorf("expected the passes to contain 3 entries")
	}

	if config.Get().Conn.Passes[2] != "pass2" {
		t.Errorf("expected the passes to be pass2 on index 2, but got %v", config.Get().Conn.Passes[2])
	}
}

func TestReadOverlayConfigReader(t *testing.T) {
	f0, _ := os.Open("testData/test_config_0.yaml")
	f1, _ := os.Open("testData/test_config_0_overlay.yaml")

	config, configErr := templig.From[TestConfig](f0, f1)

	if configErr != nil {
		t.Errorf("no error expected reading multiple files: %v", configErr)
	}

	if len(config.Get().Conn.Passes) != 3 {
		t.Errorf("expected the passes to contain 3 entries")
	}

	if config.Get().Conn.Passes[2] != "pass2" {
		t.Errorf("expected the passes to be pass2 on index 2, but got %v", config.Get().Conn.Passes[2])
	}
}

func TestReadOverlayConfigMismatch(t *testing.T) {
	_, configErr := templig.FromFile[TestConfig](
		"testData/test_config_0.yaml",
		"testData/test_config_0_overlay_mismatch.yaml",
	)

	if configErr == nil {
		t.Errorf(" error expected reading multiple incompatible files:")
	}
}

func TestReadOverlayConfigWrongType(t *testing.T) {
	_, configErr := templig.FromFile[TestConfig](
		"testData/test_config_0.yaml",
		"testData/test_config_0_overlay_wrongtype.yaml",
	)

	if configErr == nil {
		t.Errorf(" error expected reading multiple incompatible files:")
	}
}

type BrokenIO struct{}

func (b *BrokenIO) Read(_ []byte) (n int, err error) {
	return 0, errors.New("broken reader")
}

func (b *BrokenIO) Write(_ []byte) (n int, err error) {
	return 0, errors.New("broken writer")
}

func TestBrokenReader(t *testing.T) {
	c, fromErr := templig.From[TestConfig](&BrokenIO{})

	if fromErr == nil {
		t.Errorf("reading from broken reader should have returned an error")
	}

	if c != nil {
		t.Errorf("reading from broken reader should have returned nil")
	}
}

func TestReadOverlayConfigBrokenReader(t *testing.T) {
	f0 := &BrokenIO{}
	f1 := &BrokenIO{}

	c, fromErr := templig.From[TestConfig](f0, f1)

	if fromErr == nil {
		t.Errorf("reading from broken reader should have returned an error")
	}

	if c != nil {
		t.Errorf("reading from broken reader should have returned nil")
	}
}

func TestNonexistentFile(t *testing.T) {
	c, fromErr := templig.FromFile[TestConfig]("testData/test_does_not_exist.yaml")

	if fromErr == nil {
		t.Errorf("reading from broken reader should have returned an error")
	}

	if c != nil {
		t.Errorf("reading from broken reader should have returned nil")
	}
}

func TestNonexistentFileOverlayAddon(t *testing.T) {
	c, fromErr := templig.FromFile[TestConfig](
		"testData/test_config_0.yaml",
		"testData/test_does_not_exist.yaml",
	)

	if fromErr == nil {
		t.Errorf("reading from broken reader should have returned an error")
	}

	if c != nil {
		t.Errorf("reading from broken reader should have returned nil")
	}
}

func TestNoFiles(t *testing.T) {
	c, fromErr := templig.FromFile[TestConfig]([]string{}...)

	if fromErr == nil {
		t.Errorf("reading from broken reader should have returned an error")
	}

	if c != nil {
		t.Errorf("reading from broken reader should have returned nil")
	}
}

func TestNoFilesDeprecated(t *testing.T) {
	c, fromErr := templig.FromFiles[TestConfig]([]string{})

	if fromErr == nil {
		t.Errorf("reading from broken reader should have returned an error")
	}

	if c != nil {
		t.Errorf("reading from broken reader should have returned nil")
	}
}

func TestBrokenWriter(t *testing.T) {
	c, _ := templig.FromFile[TestConfig]("testData/test_config_0.yaml")

	toErr := c.To(&BrokenIO{})

	if toErr == nil {
		t.Errorf("reading from broken reader should have returned an error")
	}
}

func TestWriteFile(t *testing.T) {
	config, _ := templig.FromFile[TestConfig]("testData/test_config_0.yaml")

	err := config.ToFile("testData/test_config_written.yaml")

	if err != nil {
		t.Errorf("writing to file should work")
	}
	defer func() { _ = os.Remove("testData/test_config_written.yaml") }()

	bufOrig := bytes.Buffer{}
	bufCopy := bytes.Buffer{}

	_ = config.To(&bufOrig)

	cp, _ := templig.FromFile[TestConfig]("testData/test_config_written.yaml")
	_ = cp.To(&bufCopy)

	if bufOrig.String() != bufCopy.String() {
		t.Errorf("written file does not match original file")
	}
}

func TestWriteProtectedFile(t *testing.T) {
	c, _ := templig.FromFile[TestConfig]("testData/test_config_0.yaml")

	if chmodErr := os.Chmod("testData/test_write_protected.yaml", 0400); chmodErr != nil {
		t.Errorf("could not writeprotect file for test: %v", chmodErr)
	}

	err := c.ToFile("testData/test_write_protected.yaml")

	if err == nil {
		t.Errorf("writing to file should not work")
	}
}

func TestSecretsHidden(t *testing.T) {
	c, _ := templig.FromFile[TestConfig]("testData/test_config_0.yaml")

	buf := bytes.Buffer{}

	if err := c.ToSecretsHidden(&buf); err != nil {
		t.Errorf("could not generate secrets-hidden config")
	}

	if strings.Contains(buf.String(), "pass0") || strings.Contains(buf.String(), "pass1") {
		t.Errorf("found secrets in normally secrets-hidden output")
	}

	if !strings.Contains(buf.String(), "passes: '*'") {
		t.Errorf("did not find replaced pass secret:\n%v", buf.String())
	}
}

func TestSecretsHiddenStructured(t *testing.T) {
	c, _ := templig.FromFile[TestConfig]("testData/test_config_0.yaml")

	buf := bytes.Buffer{}

	if err := c.ToSecretsHiddenStructured(&buf); err != nil {
		t.Errorf("could not generate secrets-hidden config")
	}

	if strings.Contains(buf.String(), "pass0") || strings.Contains(buf.String(), "pass1") {
		t.Errorf("found secrets in normally secrets-hidden output")
	}

	if !strings.Contains(buf.String(), "passes:\n") {
		t.Errorf("did not find replaced pass secret:\n%v", buf.String())
	}

	if strings.Count(buf.String(), "'*****'") != 2 {
		t.Errorf("did not find replaced pass secrets:\n%v", buf.String())
	}
}

func FuzzFromFileEnv(f *testing.F) {
	f.Add("")
	f.Add("12345")
	f.Add("123456")
	f.Add("1234567")
	f.Add("pass")
	f.Add("password")
	f.Add("qwerty")
	f.Add("secret")
	f.Add("test")

	f.Fuzz(func(t *testing.T, envVar string) {
		if slices.Contains([]rune(envVar), 0) {
			return
		}

		t.Setenv("PASS1", envVar)

		_, confErr := templig.FromFile[TestConfig]("testData/test_config_1.yaml")

		if confErr != nil && len(envVar) > 0 {
			t.Errorf("got unexpected error on input -%v-: %v", envVar, confErr)
		}
	})
}

type TestConfigValidated struct {
	ID   int    `yaml:"id"`
	Name string `yaml:"name"`
}

func (c *TestConfigValidated) Validate() error {
	if c.ID == 9 {
		return nil
	}

	return fmt.Errorf("expected id 9 to be valid")
}
func TestReadConfigValidated(t *testing.T) {
	tests := []struct {
		in      string
		inFile  string
		env     map[string]string
		want    TestConfigValidated
		wantErr bool
	}{
		{ // 0
			in: `
                id:   8
                name: "Name0"`,
			want: TestConfigValidated{
				ID:   8,
				Name: "Name0",
			},
			wantErr: true,
		},
		{ // 1
			in: `
                id:   9
                name: "Name0"`,
			want: TestConfigValidated{
				ID:   9,
				Name: "Name0",
			},
			wantErr: false,
		},
	}

	testBuf := bytes.Buffer{}

	for testNum, test := range tests {
		if len(test.in) > 0 && len(test.inFile) > 0 {
			t.Errorf("%v: input data and file given at the same time", testNum)
		}

		testBuf.Reset()
		testBuf.WriteString(test.in)

		if test.env != nil {
			for ei, ev := range test.env {
				t.Setenv(ei, ev)
			}
		}

		var config *templig.Config[TestConfigValidated]
		var fromErr error

		switch {
		case len(test.in) > 0:
			config, fromErr = templig.From[TestConfigValidated](&testBuf)
		case len(test.inFile) > 0:
			config, fromErr = templig.FromFile[TestConfigValidated](test.inFile)
		default:
			t.Errorf("%v: neither input data nor input file given", testNum)
		}

		if test.wantErr && fromErr == nil {
			t.Errorf("%v: wanted error but got nil", testNum)
		}
		if !test.wantErr && fromErr != nil {
			t.Errorf("%v: did not want error but got %v", testNum, fromErr)
		}

		if config != nil {
			if config.Get().ID != test.want.ID {
				t.Errorf("%v: wanted ID %v but got %v", testNum, test.want.ID, config.Get().ID)
			}
			if config.Get().Name != test.want.Name {
				t.Errorf("%v: wanted Name %v but got %v", testNum, test.want.Name, config.Get().Name)
			}
		}

		if test.env != nil {
			for ei := range test.env {
				_ = os.Unsetenv(ei)
			}
		}
	}
}

// -----------------------------------------------------------------------------
// Additional tests (using Go's standard "testing" package) to increase coverage
// and validate edge cases, writer errors, immutability, and round-trip behavior.
// -----------------------------------------------------------------------------

// helper: strict order-equality for string slices
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestRoundTripBuffer(t *testing.T) {
	t.Parallel()

	orig, err := templig.FromFile[TestConfig]("testData/test_config_0.yaml")
	if err != nil {
		t.Fatalf("reading base config failed: %v", err)
	}

	var buf bytes.Buffer
	if err := orig.To(&buf); err != nil {
		t.Fatalf("serializing to buffer failed: %v", err)
	}

	cp, err := templig.From[TestConfig](&buf)
	if err != nil {
		t.Fatalf("deserializing from buffer failed: %v", err)
	}

	got := cp.Get()
	want := orig.Get()

	if got.ID != want.ID {
		t.Errorf("ID mismatch: want %v, got %v", want.ID, got.ID)
	}
	if got.Name != want.Name {
		t.Errorf("Name mismatch: want %q, got %q", want.Name, got.Name)
	}
	if (got.Conn == nil) != (want.Conn == nil) {
		t.Fatalf("Conn nil-mismatch: want %v, got %v", want.Conn != nil, got.Conn != nil)
	}
	if got.Conn != nil {
		if got.Conn.URL != want.Conn.URL {
			t.Errorf("URL mismatch: want %q, got %q", want.Conn.URL, got.Conn.URL)
		}
		if !equalStringSlices(got.Conn.Passes, want.Conn.Passes) {
			t.Errorf("passes mismatch: want %v, got %v", want.Conn.Passes, got.Conn.Passes)
		}
	}
}

func TestGetReturnsCopy(t *testing.T) {
	t.Parallel()

	c, err := templig.FromFile[TestConfig]("testData/test_config_0.yaml")
	if err != nil {
		t.Fatalf("reading base config failed: %v", err)
	}

	mod := c.Get()
	mod.Name = "CHANGED"

	// Ensure internal state hasn't changed
	if c.Get().Name == "CHANGED" {
		t.Errorf("Get should return a copy; internal state mutated unexpectedly")
	}

	// Also ensure serialization doesn't include the change
	var buf bytes.Buffer
	if err := c.To(&buf); err != nil {
		t.Fatalf("serializing failed: %v", err)
	}
	if strings.Contains(buf.String(), "CHANGED") {
		t.Errorf("serialized output contains modified value; expected original data only")
	}
}

func TestToSecretsHidden_WriterError(t *testing.T) {
	t.Parallel()

	c, err := templig.FromFile[TestConfig]("testData/test_config_0.yaml")
	if err != nil {
		t.Fatalf("reading base config failed: %v", err)
	}

	if err := c.ToSecretsHidden(&BrokenIO{}); err == nil {
		t.Errorf("expected error when writing secrets-hidden config to broken writer")
	}
}

func TestToSecretsHiddenStructured_WriterError(t *testing.T) {
	t.Parallel()

	c, err := templig.FromFile[TestConfig]("testData/test_config_0.yaml")
	if err != nil {
		t.Fatalf("reading base config failed: %v", err)
	}

	if err := c.ToSecretsHiddenStructured(&BrokenIO{}); err == nil {
		t.Errorf("expected error when writing structured secrets-hidden config to broken writer")
	}
}

func TestFromFile_InvalidYAML_TempFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := dir + "/invalid.yaml"
	content := "id: 23\nname: {{ required \"has to be set\" 9 | quote" // intentionally missing closing braces

	if writeErr := os.WriteFile(path, []byte(content), 0o600); writeErr != nil {
		t.Fatalf("failed writing temp invalid file: %v", writeErr)
	}

	c, err := templig.FromFile[TestConfig](path)
	if err == nil {
		t.Errorf("expected error when reading invalid YAML/template, got nil")
	}
	if c != nil {
		t.Errorf("expected returned config to be nil on invalid input")
	}
}

func TestFrom_EmptyReader(t *testing.T) {
	t.Parallel()

	var empty bytes.Buffer
	c, err := templig.From[TestConfig](&empty)
	if err == nil {
		t.Errorf("expected error when reading from empty reader, got nil")
	}
	if c != nil {
		t.Errorf("expected nil config when reading from empty reader")
	}
}

func TestSecretsHiddenStructured_RoundTripIsValidYAML(t *testing.T) {
	t.Parallel()

	c, err := templig.FromFile[TestConfig]("testData/test_config_0.yaml")
	if err != nil {
		t.Fatalf("reading base config failed: %v", err)
	}

	orig := c.Get()
	if orig.Conn == nil {
		t.Fatalf("test precondition failed: Conn should not be nil")
	}
	origCount := len(orig.Conn.Passes)

	var buf bytes.Buffer
	if err := c.ToSecretsHiddenStructured(&buf); err != nil {
		t.Fatalf("generating structured secrets-hidden failed: %v", err)
	}

	// Re-parse sanitized YAML; it should remain valid
	sanitized, err := templig.From[TestConfig](&buf)
	if err != nil {
		t.Fatalf("parsing structured secrets-hidden YAML failed: %v", err)
	}
	got := sanitized.Get()
	if got.Conn == nil {
		t.Fatalf("expected Conn not to be nil in sanitized output")
	}
	if len(got.Conn.Passes) != origCount {
		t.Errorf("expected %d sanitized passes, got %d", origCount, len(got.Conn.Passes))
	}
	for i, p := range got.Conn.Passes {
		if p != "*****" {
			t.Errorf("expected pass %d to be '*****', got %q", i, p)
		}
	}
	// Ensure no original secrets leaked into the sanitized text
	out := buf.String()
	for _, secret := range orig.Conn.Passes {
		if strings.Contains(out, secret) {
			t.Errorf("found leaked secret %q in sanitized output", secret)
		}
	}
}

func TestReadOverlayConfig_ValidThenBrokenReader(t *testing.T) {
	t.Parallel()

	f0, openErr := os.Open("testData/test_config_0.yaml")
	if openErr != nil {
		t.Fatalf("failed to open base config: %v", openErr)
	}
	defer func() { _ = f0.Close() }()

	if c, err := templig.From[TestConfig](f0, &BrokenIO{}); err == nil || c != nil {
		t.Errorf("expected error and nil config when one of the overlay readers is broken")
	}
}

func TestSecretsHidden_DoesNotMutateOriginal(t *testing.T) {
	t.Parallel()

	c, err := templig.FromFile[TestConfig]("testData/test_config_0.yaml")
	if err != nil {
		t.Fatalf("reading base config failed: %v", err)
	}
	before := c.Get()

	var buf bytes.Buffer
	if err := c.ToSecretsHidden(&buf); err != nil {
		t.Fatalf("generating secrets-hidden failed: %v", err)
	}

	after := c.Get()
	if before.ID != after.ID || before.Name != after.Name {
		t.Errorf("unexpected mutation of scalar fields: before=%v after=%v", before, after)
	}
	switch {
	case (before.Conn == nil) != (after.Conn == nil):
		t.Errorf("unexpected mutation of Conn presence")
	case before.Conn != nil:
		if before.Conn.URL != after.Conn.URL {
			t.Errorf("unexpected mutation of Conn.URL: before=%q after=%q", before.Conn.URL, after.Conn.URL)
		}
		if !equalStringSlices(before.Conn.Passes, after.Conn.Passes) {
			t.Errorf("unexpected mutation of Conn.Passes: before=%v after=%v", before.Conn.Passes, after.Conn.Passes)
		}
	}
}

func TestSecretsHidden_PreservesNonSecretFields(t *testing.T) {
	t.Parallel()

	c, err := templig.FromFile[TestConfig]("testData/test_config_0.yaml")
	if err != nil {
		t.Fatalf("reading base config failed: %v", err)
	}

	var buf bytes.Buffer
	if err := c.ToSecretsHidden(&buf); err != nil {
		t.Fatalf("generating secrets-hidden failed: %v", err)
	}

	out := buf.String()
	// Ensure a known non-secret field remains visible
	if !strings.Contains(out, "url: https://www.tests.to") {
		t.Errorf("expected non-secret URL to be preserved in secrets-hidden output; got:\n%s", out)
	}
}
