class OllamaLoggingProxyCanary < Formula
  desc "Reverse proxy in front of Ollama with JSONL request and response logging (canary prerelease)"
  homepage "https://github.com/josephma93/ollama-debug-logging-proxy"
  version "0.1.0-canary.1"
  url "https://github.com/josephma93/ollama-debug-logging-proxy/archive/refs/tags/v0.1.0-canary.1.tar.gz"
  sha256 "57bf5a651b40d96cc2cef7254dffc10290c2e65103659c987241aeee08b714ef"
  conflicts_with "ollama-logging-proxy", because: "both formulae install the same proxy and helper command names"
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
      This canary formula tracks prerelease tags and conflicts with the stable formula.

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
