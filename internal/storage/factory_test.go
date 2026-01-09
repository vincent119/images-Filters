package storage

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vincent119/images-filters/internal/config"
)

func TestNewStorage(t *testing.T) {
	tmpDir, _ := ioutil.TempDir("", "storage_test")
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name        string
		cfg         *config.Config
		wantErr     bool
		expectedType interface{} // We can check type if we export concrete types or use reflection
	}{
		{
			name: "Local Storage - Success",
			cfg: &config.Config{
				Storage: config.StorageConfig{
					Type: "local",
					Local: config.LocalStorageConfig{
						RootPath: tmpDir,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Local Storage - Missing Root",
			cfg: &config.Config{
				Storage: config.StorageConfig{
					Type: "local",
					Local: config.LocalStorageConfig{
						RootPath: "",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "No Storage",
			cfg: &config.Config{
				Storage: config.StorageConfig{
					Type: "no_storage",
				},
			},
			wantErr: false,
		},
		{
			name: "Unsupported Type",
			cfg: &config.Config{
				Storage: config.StorageConfig{
					Type: "unknown",
				},
			},
			wantErr: true,
		},
		{
			name: "Mixed Storage - Valid",
			cfg: &config.Config{
				Storage: config.StorageConfig{
					Type: "mixed",
					Mixed: config.MixedStorageConfig{
						SourceStorage: "local",
						ResultStorage: "no_storage",
					},
					Local: config.LocalStorageConfig{
						RootPath: tmpDir,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Mixed Storage - Nested Mixed Error",
			cfg: &config.Config{
				Storage: config.StorageConfig{
					Type: "mixed",
					Mixed: config.MixedStorageConfig{
						SourceStorage: "mixed",
						ResultStorage: "local",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Mixed Storage - Source Error",
			cfg: &config.Config{
				Storage: config.StorageConfig{
					Type: "mixed",
					Mixed: config.MixedStorageConfig{
						SourceStorage: "unknown",
						ResultStorage: "local",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "Mixed Storage - Result Error",
			cfg: &config.Config{
				Storage: config.StorageConfig{
					Type: "mixed",
					Mixed: config.MixedStorageConfig{
						SourceStorage: "local",
						ResultStorage: "unknown",
					},
					Local: config.LocalStorageConfig{
						RootPath: tmpDir,
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := NewStorage(context.Background(), tt.cfg)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, s)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, s)
			}
		})
	}
}
