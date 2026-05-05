class OllamaLoggingProxy < Formula
  desc "Reverse proxy in front of Ollama with JSONL request and response logging"
  homepage "https://github.com/josephma93/ollama-debug-logging-proxy"
  version "{{VERSION}}"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/josephma93/ollama-debug-logging-proxy/releases/download/{{TAG}}/ollama-logging-proxy-aarch64-apple-darwin.tar.gz"
      sha256 "{{SHA_ARM64}}"
    else
      url "https://github.com/josephma93/ollama-debug-logging-proxy/releases/download/{{TAG}}/ollama-logging-proxy-x86_64-apple-darwin.tar.gz"
      sha256 "{{SHA_X64}}"
    end
  end

  def install
    bin.install "ollama-logging-proxy"
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

  test do
    output = shell_output("#{bin}/ollama-logging-proxy no-such-command 2>&1", 1)
    assert_match "unknown command", output
  end
end
