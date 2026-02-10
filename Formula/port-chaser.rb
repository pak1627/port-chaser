# typed: false
# frozen_string_literal: true

# Formula: port-chaser
# Description: Terminal UI-based port management tool for developers
# Homepage: https://github.com/manson/port-chaser
# License: MIT

class PortChaser < Formula
  desc "Terminal UI-based port management tool for developers"
  homepage "https://github.com/manson/port-chaser"
  url "https://github.com/manson/port-chaser/archive/refs/tags/v0.1.0.tar.gz"
  sha256 "PLACEHOLDER_SHA256"
  license "MIT"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w"), "./cmd/port-chaser"
  end

  test do
    assert_match("Port Chaser", shell_output("#{bin}/port-chaser --version"))
    assert_match("0.1.0", shell_output("#{bin}/port-chaser --version"))
  end
end
