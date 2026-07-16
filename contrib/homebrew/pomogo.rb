class Pomogo < Formula
  desc "A beautiful terminal deep-work companion for developers."
  homepage "https://github.com/Ibnu-Afdel/pomogo"
  url "https://github.com/Ibnu-Afdel/pomogo/releases/download/v2.0.0/pomogo-2.0.0-linux-amd64.tar.gz"
  sha256 "SKIP"
  license "MIT"

  def install
    bin.install "pomogo"
  end

  test do
    system "#{bin}/pomogo", "version"
  end
end
