class OllamaLoggingProxy < Formula
  desc "Reverse proxy in front of Ollama with JSONL request and response logging"
  homepage "https://github.com/josephma93/ollama-debug-logging-proxy"
  disable! date: "2026-05-05", because: "no stable release has been published yet; use the canary formula until the first stable tag exists"
end
