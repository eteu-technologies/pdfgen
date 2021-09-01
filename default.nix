{ lib, buildGoModule, runCommandNoCC, git, rev ? null }:

let
  versionInfo = src: import (runCommandNoCC "eteu-pdfgen-version" { } ''
    v=$(${git}/bin/git -C "${src}" rev-parse HEAD || echo "0000000000000000000000000000000000000000")
    printf '{ version = "%s"; }' "$v" > $out
  '');

  # Need to keep .git around for version string
  srcCleaner = name: type: let baseName = baseNameOf (toString name); in (baseName == ".git" || lib.cleanSourceFilter name type);
in
buildGoModule rec {
  pname = "eteu-pdfgen";
  version = if (rev != null) then rev else (versionInfo src).version;

  src = lib.cleanSourceWith { filter = srcCleaner; src = ./.; };

  ldflags = [
    "-X github.com/eteu-technologies/pdfgen/internal/core.Version=${version}"
  ];

  doCheck = true;

  vendorSha256 = "sha256-v8cWnkCNP4PAT6Bz/8GmY2EUysgTI6XTCfzhXuumvak=";
  subPackages = [ "cmd/pdfgen" ];
}
