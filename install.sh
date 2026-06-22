#!/bin/sh
set -eu

repo="prilog-ai/prilog-cli"
binary="prilog"
install_dir="${PRILOG_INSTALL_DIR:-/usr/local/bin}"
version="${PRILOG_VERSION:-latest}"

say() {
	printf '%s\n' "$*"
}

fail() {
	printf 'prilog install: %s\n' "$*" >&2
	exit 1
}

download() {
	url="$1"
	out="$2"

	if command -v curl >/dev/null 2>&1; then
		curl -fsSL "$url" -o "$out"
		return
	fi

	if command -v wget >/dev/null 2>&1; then
		wget -qO "$out" "$url"
		return
	fi

	fail "curl or wget is required"
}

os_name="$(uname -s 2>/dev/null || true)"
arch_name="$(uname -m 2>/dev/null || true)"

case "$os_name" in
	Darwin) os="darwin" ;;
	Linux) os="linux" ;;
	*) fail "unsupported operating system: ${os_name:-unknown}" ;;
esac

case "$arch_name" in
	x86_64 | amd64) arch="amd64" ;;
	arm64 | aarch64) arch="arm64" ;;
	*) fail "unsupported architecture: ${arch_name:-unknown}" ;;
esac

asset="${binary}_${os}_${arch}.tar.gz"
if [ "$version" = "latest" ]; then
	base_url="https://github.com/${repo}/releases/latest/download"
else
	base_url="https://github.com/${repo}/releases/download/${version}"
fi

if command -v mktemp >/dev/null 2>&1; then
	tmp_dir="$(mktemp -d "${TMPDIR:-/tmp}/prilog-install.XXXXXX")"
else
	tmp_dir="${TMPDIR:-/tmp}/prilog-install.$$"
	mkdir -p "$tmp_dir"
fi
archive_path="${tmp_dir}/${asset}"
checksums_path="${tmp_dir}/checksums.txt"

cleanup() {
	rm -rf "$tmp_dir"
}
trap cleanup EXIT INT TERM

say "Installing Prilog CLI"
say "Platform: ${os}/${arch}"
say "Source: ${base_url}/${asset}"

download "${base_url}/${asset}" "$archive_path" || fail "could not download ${asset}"

if download "${base_url}/checksums.txt" "$checksums_path" >/dev/null 2>&1; then
	expected="$(grep " ${asset}$" "$checksums_path" | awk '{print $1}' || true)"
	if [ -n "$expected" ]; then
		actual=""
		if command -v sha256sum >/dev/null 2>&1; then
			actual="$(sha256sum "$archive_path" | awk '{print $1}')"
		elif command -v shasum >/dev/null 2>&1; then
			actual="$(shasum -a 256 "$archive_path" | awk '{print $1}')"
		fi

		if [ -n "$actual" ]; then
			[ "$actual" = "$expected" ] || fail "checksum verification failed"
			say "Checksum verified"
		fi
	fi
fi

tar -xzf "$archive_path" -C "$tmp_dir"
binary_path="$(find "$tmp_dir" -type f -name "$binary" | head -n 1)"
[ -n "$binary_path" ] || fail "archive did not contain ${binary}"
chmod 0755 "$binary_path"

if [ -w "$install_dir" ]; then
	cp "$binary_path" "${install_dir}/${binary}"
	chmod 0755 "${install_dir}/${binary}"
else
	if command -v sudo >/dev/null 2>&1; then
		say "Installing to ${install_dir} with sudo"
		sudo mkdir -p "$install_dir"
		sudo cp "$binary_path" "${install_dir}/${binary}"
		sudo chmod 0755 "${install_dir}/${binary}"
	else
		[ -n "${HOME:-}" ] || fail "${install_dir} is not writable and HOME is not set"
		install_dir="${HOME}/.local/bin"
		mkdir -p "$install_dir"
		cp "$binary_path" "${install_dir}/${binary}"
		chmod 0755 "${install_dir}/${binary}"
	fi
fi

case ":${PATH}:" in
	*":${install_dir}:"*) ;;
	*) say "Add ${install_dir} to PATH before running ${binary}" ;;
esac

installed_version="$("${install_dir}/${binary}" version 2>/dev/null || true)"
if [ -n "$installed_version" ]; then
	say "Installed ${installed_version} at ${install_dir}/${binary}"
else
	say "Installed ${binary} at ${install_dir}/${binary}"
fi

say "Next: run 'prilog login' or 'prilog init' in your repository"
