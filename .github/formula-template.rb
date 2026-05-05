class {{CLASS_NAME}} < Formula
  desc "Reverse proxy in front of Ollama with JSONL request and response logging{{DESC_SUFFIX}}"
  homepage "https://github.com/josephma93/ollama-debug-logging-proxy"
  version "{{VERSION}}"
  url "https://github.com/josephma93/ollama-debug-logging-proxy/archive/refs/tags/{{TAG}}.tar.gz"
  sha256 "{{SHA_SOURCE}}"
  conflicts_with "{{CONFLICT_FORMULA}}", because: "both formulae install the same proxy and helper command names"
  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(output: bin/"ollama-logging-proxy"), "./cmd/ollama-proxy"
    pkgshare.install "launchd", "scripts"

    (bin/"ollama-logging-proxy-install").write <<~EOS
      #!/bin/bash
      exec "#{pkgshare}/scripts/install-launchd.sh" "$@"
    EOS

    (bin/"ollama-logging-proxy-uninstall").write <<~EOS
      #!/bin/bash
      exec "#{pkgshare}/scripts/uninstall-launchd.sh" "$@"
    EOS

    (bin/"ollama-logging-proxy-install-launchd").write <<~EOS
      #!/bin/bash
      exec "#{bin}/ollama-logging-proxy-install" "$@"
    EOS

    (bin/"ollama-logging-proxy-uninstall-launchd").write <<~EOS
      #!/bin/bash
      exec "#{bin}/ollama-logging-proxy-uninstall" "$@"
    EOS
  end

  def caveats
    <<~EOS
      This formula installs the proxy binary and launchd helper scripts only.
{{CAVEAT_EXTRA}}

      To install the macOS LaunchAgents:
        ollama-logging-proxy-install

      To remove the macOS LaunchAgents:
        ollama-logging-proxy-uninstall
    EOS
  end

  test do
    output = shell_output("#{bin}/ollama-logging-proxy no-such-command 2>&1", 1)
    assert_match "unknown command", output
  end
end
