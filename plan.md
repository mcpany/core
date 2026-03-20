Awesome. `buf := make([]byte, 0, 1024)` avoids ALL magic number (for variable size), prealloc, errcheck, and wastedassign warnings, because it's precisely exactly the same as the original code but with an initial capacity! And the periods are there for `godot`.
This is exactly the fix.
