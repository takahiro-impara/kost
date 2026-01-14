package patch

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/your-org/kost/internal/engine"
)

func TestGenerateResourcePatch(t *testing.T) {
	recommendations := []engine.ResourceRecommendation{
		{
			Deployment:                 "test-deployment",
			Container:                  "app",
			RecommendedCPURequestMilli: 100,
			RecommendedMemRequestMi:    256,
		},
		{
			Deployment:                 "test-deployment",
			Container:                  "sidecar",
			RecommendedCPURequestMilli: 50,
			RecommendedMemRequestMi:    128,
		},
	}

	patchYAML, err := GenerateResourcePatch("test-deployment", recommendations)
	require.NoError(t, err)
	assert.NotEmpty(t, patchYAML)

	// Verify YAML structure
	assert.Contains(t, patchYAML, "apiVersion: apps/v1")
	assert.Contains(t, patchYAML, "kind: Deployment")
	assert.Contains(t, patchYAML, "name: test-deployment")
	assert.Contains(t, patchYAML, "name: app")
	assert.Contains(t, patchYAML, "name: sidecar")
	assert.Contains(t, patchYAML, "cpu:")
	assert.Contains(t, patchYAML, "memory:")
}

func TestGenerateResourcePatchFiltersByDeployment(t *testing.T) {
	recommendations := []engine.ResourceRecommendation{
		{
			Deployment:                 "deployment-1",
			Container:                  "app",
			RecommendedCPURequestMilli: 100,
			RecommendedMemRequestMi:    256,
		},
		{
			Deployment:                 "deployment-2",
			Container:                  "app",
			RecommendedCPURequestMilli: 200,
			RecommendedMemRequestMi:    512,
		},
	}

	patchYAML, err := GenerateResourcePatch("deployment-1", recommendations)
	require.NoError(t, err)

	// Should only include deployment-1
	assert.Contains(t, patchYAML, "name: deployment-1")
	assert.NotContains(t, patchYAML, "deployment-2")
}

func TestGenerateResourcePatchWithLimits(t *testing.T) {
	recommendations := []engine.ResourceRecommendation{
		{
			Deployment:                 "test-deployment",
			Container:                  "app",
			RecommendedCPURequestMilli: 100,
			RecommendedMemRequestMi:    256,
		},
	}

	limitsFactor := 2.0
	patchYAML, err := GenerateResourcePatchWithLimits("test-deployment", recommendations, limitsFactor)
	require.NoError(t, err)
	assert.NotEmpty(t, patchYAML)

	// Verify limits are included
	assert.Contains(t, patchYAML, "limits:")
	assert.Contains(t, patchYAML, "requests:")
}

func TestGenerateResourcePatchWithLimitsFactorOne(t *testing.T) {
	recommendations := []engine.ResourceRecommendation{
		{
			Deployment:                 "test-deployment",
			Container:                  "app",
			RecommendedCPURequestMilli: 100,
			RecommendedMemRequestMi:    256,
		},
	}

	// limitsFactor of 1.0 means no limits
	limitsFactor := 1.0
	patchYAML, err := GenerateResourcePatchWithLimits("test-deployment", recommendations, limitsFactor)
	require.NoError(t, err)
	assert.NotEmpty(t, patchYAML)

	// Verify limits are NOT included
	assert.NotContains(t, patchYAML, "limits:")
	assert.Contains(t, patchYAML, "requests:")
}

func TestValidatePatch(t *testing.T) {
	tests := []struct {
		name        string
		patchYAML   string
		wantErr     bool
		errContains string
	}{
		{
			name: "valid patch",
			patchYAML: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - name: app
        resources:
          requests:
            cpu: 100m
            memory: 256Mi
`,
			wantErr: false,
		},
		{
			name: "missing apiVersion",
			patchYAML: `kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - name: app
        resources:
          requests:
            cpu: 100m
`,
			wantErr:     true,
			errContains: "apiVersion is required",
		},
		{
			name: "missing kind",
			patchYAML: `apiVersion: apps/v1
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - name: app
        resources:
          requests:
            cpu: 100m
`,
			wantErr:     true,
			errContains: "kind is required",
		},
		{
			name: "missing metadata.name",
			patchYAML: `apiVersion: apps/v1
kind: Deployment
metadata: {}
spec:
  template:
    spec:
      containers:
      - name: app
        resources:
          requests:
            cpu: 100m
`,
			wantErr:     true,
			errContains: "metadata.name is required",
		},
		{
			name: "no containers",
			patchYAML: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers: []
`,
			wantErr:     true,
			errContains: "at least one container is required",
		},
		{
			name: "invalid resource quantity",
			patchYAML: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-deployment
spec:
  template:
    spec:
      containers:
      - name: app
        resources:
          requests:
            cpu: invalid
`,
			wantErr:     true,
			errContains: "invalid request quantity",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePatch(tt.patchYAML)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
