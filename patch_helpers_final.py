import re

with open("server/tests/integration/e2e_helpers.go", "r") as f:
    content = f.read()

content = content.replace("var roots []string", "roots := make([]string, 0)")
content = content.replace("""func symlinkIfPresent(src, dst string) error {
	if _, err := os.Stat(src); err != nil {
		return err
	}
	return os.Symlink(src, dst)
}""", """func symlinkIfPresent(src, dst string) error {
	if _, err := os.Stat(src); err != nil {
		return nil //nolint:nilerr
	}
	return os.Symlink(src, dst)
}""")

with open("server/tests/integration/e2e_helpers.go", "w") as f:
    f.write(content)
