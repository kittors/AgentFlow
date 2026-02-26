"""Shared test fixtures."""

import sys
from pathlib import Path

# Ensure project root is on sys.path
project_root = Path(__file__).parent.parent
if str(project_root) not in sys.path:
    sys.path.insert(0, str(project_root))
