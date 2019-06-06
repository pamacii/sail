package environment

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.coder.com/sail/internal/randstr"
)

func Test_Modifier(t *testing.T) {
	checkDockerDaemon(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	createBaseEnv := func(t *testing.T) (env *Environment, destroy func(*testing.T)) {
		// TODO: Remove reliance on needing to have a repo to create an env.
		name := "modifier-test-" + randstr.MakeCharset(randstr.Lower, 5)
		t.Logf("created base env: %s", name)
		cfg := NewDefaultBuildConfig(name)

		env, err := Build(ctx, cfg)
		require.NoError(t, err)

		destroy = func(t *testing.T) {
			err := Purge(ctx, env)
			require.NoError(t, err)
		}

		return env, destroy
	}

	requireEnvExec := func(t *testing.T, env *Environment, cmdStr string, args ...string) {
		out, err := env.Exec(ctx, cmdStr, args...).CombinedOutput()
		require.NoError(t, err, "failed to run command: %s", out)
	}

	t.Run("Simple", func(t *testing.T) {
		env, destroy := createBaseEnv(t)
		defer destroy(t)

		provider := RawDockerfileProvider([]byte(`FROM codercom/ubuntu-dev
			RUN sudo apt update -y && sudo apt install fish -y`))

		env, err := Modify(ctx, provider, env)
		require.NoError(t, err)

		// Simple check to see if fish was installed.
		requireEnvExec(t, env, "fish", "-c", "echo hello")

		t.Run("Stacked", func(t *testing.T) {
			provider = RawDockerfileProvider([]byte(`FROM codercom/ubuntu-dev
				RUN sudo mkdir /test-dir`))

			env, err := Modify(ctx, provider, env)
			require.NoError(t, err)

			// Fish should still be installed.
			requireEnvExec(t, env, "fish", "-c", "echo hello")

			// And the dir should exist.
			requireEnvExec(t, env, "ls", "/test-dir")
		})
	})
}