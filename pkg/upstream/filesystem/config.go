
package filesystem

import (
	"errors"
	"fmt"
	"os"

	"github.com/mcpany/core/proto/config/v1"
)

func Validate(cfg *v1.FileSystemServiceConfig) error {
	if cfg.GetBasePath() == "" {
		return errors.New("base path is required")
	}

	info, err := os.Stat(cfg.GetBasePath())
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("base path does not exist: %s", cfg.GetBasePath())
		}
		return fmt.Errorf("failed to stat base path: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("base path is not a directory: %s", cfg.GetBasePath())
	}

	return nil
}
