# Missing __init__.py

## What this is

This signal detects Python package directories that contain `.py` files but are missing the `__init__.py` file. In Python, a directory must contain an `__init__.py` file (even if empty) to be recognized as a package that can be imported.

While Python 3.3+ introduced namespace packages that don't require `__init__.py`, most projects still use regular packages, and missing `__init__.py` files cause import failures.

## Why this matters

**Import Failures**:
- **ModuleNotFoundError**: Code that tries to import from the package will fail
- **Build failures**: Package installation and distribution breaks
- **Test failures**: Test discovery may not find tests in the directory
- **IDE confusion**: Code editors may not recognize the directory as a package

**Development Impact**:
- **Wasted debugging time**: Import errors can be confusing and hard to diagnose
- **CI/CD failures**: Automated tests fail in unexpected ways
- **Deployment issues**: Production deployments may fail if imports break

**Example**:
```
myproject/
├── mypackage/          # Missing __init__.py
│   ├── module1.py
│   └── module2.py
└── main.py

# In main.py:
from mypackage import module1  # ModuleNotFoundError!
```

## How to remediate

### Add __init__.py to package directories

**Find directories missing __init__.py**:
```bash
# Find Python package directories without __init__.py
find . -type f -name "*.py" -not -name "__init__.py" -not -name "setup.py" \
  -exec dirname {} \; | sort -u | while read dir; do
  if [ ! -f "$dir/__init__.py" ]; then
    echo "$dir"
  fi
done
```

**Create __init__.py files**:
```bash
# Create empty __init__.py in a specific directory
touch mypackage/__init__.py

# Or create in all package directories
find . -type f -name "*.py" -not -name "__init__.py" -not -name "setup.py" \
  -exec dirname {} \; | sort -u | while read dir; do
  if [ ! -f "$dir/__init__.py" ]; then
    touch "$dir/__init__.py"
    echo "Created $dir/__init__.py"
  fi
done
```

### What to put in __init__.py

**Option 1: Empty file** (most common):
```python
# mypackage/__init__.py
# (empty file is fine)
```

**Option 2: Import submodules** for convenience:
```python
# mypackage/__init__.py
from .module1 import MyClass
from .module2 import my_function

__all__ = ['MyClass', 'my_function']
```

**Option 3: Package metadata**:
```python
# mypackage/__init__.py
"""
MyPackage - A Python package for doing things.
"""

__version__ = '1.0.0'
__author__ = 'Your Name'

from .module1 import MyClass
from .module2 import my_function
```

**Option 4: Lazy imports** for large packages:
```python
# mypackage/__init__.py
def __getattr__(name):
    if name == 'MyClass':
        from .module1 import MyClass
        return MyClass
    raise AttributeError(f"module {__name__!r} has no attribute {name!r}")
```

### Verify imports work

**Test imports**:
```bash
# Test that the package can be imported
python3 -c "import mypackage; print('Success!')"

# Test specific imports
python3 -c "from mypackage import module1; print('Success!')"
```

**Run tests**:
```bash
# Run pytest to verify test discovery works
pytest

# Or run specific tests
python3 -m unittest discover
```

### Platform-specific considerations

**Windows**:
```powershell
# Find directories missing __init__.py
Get-ChildItem -Recurse -Filter "*.py" |
  Where-Object { $_.Name -ne "__init__.py" -and $_.Name -ne "setup.py" } |
  ForEach-Object { $_.DirectoryName } |
  Select-Object -Unique |
  Where-Object { -not (Test-Path "$_\__init__.py") }

# Create __init__.py files
Get-ChildItem -Recurse -Filter "*.py" |
  Where-Object { $_.Name -ne "__init__.py" } |
  ForEach-Object { $_.DirectoryName } |
  Select-Object -Unique |
  Where-Object { -not (Test-Path "$_\__init__.py") } |
  ForEach-Object { New-Item -Path "$_\__init__.py" -ItemType File }
```

**macOS/Linux**:
```bash
# Create __init__.py in all package directories
find . -type d -not -path "*/\.*" -not -path "*/venv/*" \
  -not -path "*/node_modules/*" | while read dir; do
  if ls "$dir"/*.py 2>/dev/null | grep -v __init__.py | grep -q .; then
    if [ ! -f "$dir/__init__.py" ]; then
      touch "$dir/__init__.py"
      echo "Created $dir/__init__.py"
    fi
  fi
done
```

### Best practices

1. **Always create __init__.py** when creating a new package directory:
   ```bash
   mkdir mypackage
   touch mypackage/__init__.py
   ```

2. **Use a project template** that includes __init__.py:
   ```bash
   # Use cookiecutter or similar
   cookiecutter gh:audreyr/cookiecutter-pypackage
   ```

3. **Add to your pre-commit hook**:
   ```bash
   #!/bin/bash
   # .git/hooks/pre-commit

   # Find Python packages missing __init__.py
   missing=$(find . -type f -name "*.py" -not -name "__init__.py" \
     -exec dirname {} \; | sort -u | while read dir; do
     if [ ! -f "$dir/__init__.py" ]; then
       echo "$dir"
     fi
   done)

   if [ -n "$missing" ]; then
     echo "Error: Python packages missing __init__.py:"
     echo "$missing"
     exit 1
   fi
   ```

4. **Use IDE features** to create packages:
   - PyCharm: Right-click → New → Python Package (creates __init__.py automatically)
   - VS Code: Use Python extension's "New Python Package" command

5. **Document package structure** in your README:
   ```markdown
   ## Project Structure
   ```
   myproject/
   ├── mypackage/
   │   ├── __init__.py
   │   ├── module1.py
   │   └── module2.py
   └── tests/
       ├── __init__.py
       └── test_module1.py
   ```
   ```

### When __init__.py is not needed

**Python 3.3+ namespace packages**:
```python
# If you're explicitly using namespace packages (PEP 420)
# you don't need __init__.py, but this is rare
```

**Script directories** (not packages):
```bash
# If the directory contains standalone scripts, not a package
scripts/
├── script1.py  # Standalone script
└── script2.py  # Standalone script
# No __init__.py needed here
```


## Performance

This signal uses several optimizations to maintain the 10ms execution target:

**Project Detection Gate**: The signal only runs if the current directory looks like a Python project (contains `setup.py`, `pyproject.toml`, `requirements.txt`, or `.py` files in the root). This prevents expensive directory scans in non-Python directories.

**Traversal Limits**:
- Maximum depth of 6 levels (Python packages rarely go deeper)
- Maximum of 500 directories visited
- Early exit after finding the first missing `__init__.py`

If you're in a directory that contains Python projects but isn't itself a Python project (like `~/repos`), the signal will skip silently rather than scanning all subdirectories.

## Disabling This Signal

To disable this signal, set the environment variable:
```
export DASHLIGHTS_DISABLE_MISSING_INIT_PY=1
```

To disable permanently, add the above line to your shell configuration file (`~/.zshrc`, `~/.bashrc`, etc.).
