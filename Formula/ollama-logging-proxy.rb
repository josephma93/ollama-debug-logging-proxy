class OllamaLoggingProxy < Formula
  desc "Reverse proxy in front of Ollama with JSONL request and response logging"
  homepage "https://github.com/josephma93/ollama-debug-logging-proxy"
  version "0.0.1"
  url "https://github.com/josephma93/ollama-debug-logging-proxy/archive/refs/tags/v0.0.1.tar.gz"
  sha256 "0b3e04758b7fe7ed5a8f164f26034fed48e0d01a16502621347f199e6882159e"
  conflicts_with "ollama-logging-proxy-canary", because: "both formulae install the same proxy and helper command names"
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
