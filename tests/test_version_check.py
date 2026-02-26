"""Tests for version_check.py — cache read/write, fetch, check_update."""

import json
import time
from unittest import mock

# ── Cache operations ──────────────────────────────────────────────────────────


class TestCache:
    def test_read_cache_empty(self, tmp_path):
        from agentflow.version_check import _read_cache

        with (
            mock.patch("agentflow.version_check._CACHE_FILE", tmp_path / "nonexistent.json"),
        ):
            assert _read_cache() is None

    def test_write_and_read_cache(self, tmp_path):
        from agentflow.version_check import _read_cache, _write_cache

        cache_dir = tmp_path / "cache"
        cache_file = cache_dir / "version_cache.json"

        with (
            mock.patch("agentflow.version_check._CACHE_DIR", cache_dir),
            mock.patch("agentflow.version_check._CACHE_FILE", cache_file),
        ):
            _write_cache({"latest": "2.0.0", "timestamp": time.time()})
            data = _read_cache()
            assert data is not None
            assert data["latest"] == "2.0.0"

    def test_read_cache_corrupt(self, tmp_path):
        from agentflow.version_check import _read_cache

        cache_file = tmp_path / "corrupt.json"
        cache_file.write_text("not json{{{")

        with mock.patch("agentflow.version_check._CACHE_FILE", cache_file):
            assert _read_cache() is None


# ── _fetch_latest_version ─────────────────────────────────────────────────────


class TestFetchLatestVersion:
    def test_fetch_success(self):
        from agentflow.version_check import _fetch_latest_version

        mock_response = mock.Mock()
        mock_response.read.return_value = json.dumps({"tag_name": "v2.0.0"}).encode()
        mock_response.__enter__ = mock.Mock(return_value=mock_response)
        mock_response.__exit__ = mock.Mock(return_value=False)

        with mock.patch("urllib.request.urlopen", return_value=mock_response):
            result = _fetch_latest_version()
            assert result == "2.0.0"

    def test_fetch_strips_v_prefix(self):
        from agentflow.version_check import _fetch_latest_version

        mock_response = mock.Mock()
        mock_response.read.return_value = json.dumps({"tag_name": "v1.2.3"}).encode()
        mock_response.__enter__ = mock.Mock(return_value=mock_response)
        mock_response.__exit__ = mock.Mock(return_value=False)

        with mock.patch("urllib.request.urlopen", return_value=mock_response):
            result = _fetch_latest_version()
            assert result == "1.2.3"

    def test_fetch_failure(self):
        from agentflow.version_check import _fetch_latest_version

        with mock.patch("urllib.request.urlopen", side_effect=Exception("timeout")):
            assert _fetch_latest_version() is None


# ── check_update ──────────────────────────────────────────────────────────────


class TestCheckUpdate:
    def test_show_version(self, capsys):
        from agentflow.version_check import check_update

        with mock.patch("importlib.metadata.version", return_value="1.0.0"):
            check_update(show_version=True, cache_ttl_hours=0)
        out = capsys.readouterr().out
        assert "1.0.0" in out

    def test_ttl_zero_skips(self, capsys):
        """cache_ttl_hours=0 should return immediately."""
        from agentflow.version_check import check_update

        with mock.patch("importlib.metadata.version", return_value="1.0.0"):
            check_update(cache_ttl_hours=0)
        out = capsys.readouterr().out
        # Should not fetch or print update info
        assert "⬆️" not in out

    def test_cache_hit_same_version(self, capsys, tmp_path):
        """Cached version = current → no update message."""
        from agentflow.version_check import check_update

        cache_file = tmp_path / "version_cache.json"
        cache_file.write_text(
            json.dumps(
                {
                    "latest": "1.0.0",
                    "timestamp": time.time(),
                }
            )
        )

        with (
            mock.patch("importlib.metadata.version", return_value="1.0.0"),
            mock.patch("agentflow.version_check._CACHE_FILE", cache_file),
        ):
            check_update(cache_ttl_hours=24)
        out = capsys.readouterr().out
        assert "⬆️" not in out

    def test_cache_hit_new_version(self, capsys, tmp_path):
        """Cached version > current → show update message."""
        from agentflow.version_check import check_update

        cache_file = tmp_path / "version_cache.json"
        cache_file.write_text(
            json.dumps(
                {
                    "latest": "2.0.0",
                    "timestamp": time.time(),
                }
            )
        )

        with (
            mock.patch("importlib.metadata.version", return_value="1.0.0"),
            mock.patch("agentflow.version_check._CACHE_FILE", cache_file),
        ):
            check_update(cache_ttl_hours=24)
        out = capsys.readouterr().out
        assert "2.0.0" in out

    def test_force_bypasses_cache(self, tmp_path):
        """force=True should fetch even with valid cache."""
        from agentflow.version_check import check_update

        cache_file = tmp_path / "version_cache.json"
        cache_dir = tmp_path / "cache_dir"
        cache_file.write_text(
            json.dumps(
                {
                    "latest": "1.0.0",
                    "timestamp": time.time(),
                }
            )
        )

        with (
            mock.patch("importlib.metadata.version", return_value="1.0.0"),
            mock.patch("agentflow.version_check._CACHE_FILE", cache_file),
            mock.patch("agentflow.version_check._CACHE_DIR", cache_dir),
            mock.patch("agentflow.version_check._fetch_latest_version", return_value="1.0.0") as m,
        ):
            check_update(force=True)
            m.assert_called_once()
