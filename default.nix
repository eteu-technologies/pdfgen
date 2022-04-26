{ lib, buildGoModule, rev ? "dirty" }:

buildGoModule rec {
  pname = "eteu-pdfgen";
  version = rev;

  src = lib.cleanSource ./.;

  ldflags = [
    "-X github.com/eteu-technologies/pdfgen/internal/core.Version=${version}"
  ];

  doCheck = true;

  vendorSha256 = "sha256-11Ur5HJ+PGGJV7iYCbnnU13QUBvU5ib05Cjb+TrMZEw=";
  subPackages = [ "cmd/pdfgen" ];
}
