{ lib, buildGoModule, go_1_17, runCommandNoCC, git, rev ? null }:

let
  versionInfo = src: import (runCommandNoCC "eteu-pdfgen-version" { } ''
    v=$(${git}/bin/git -C "${src}" rev-parse HEAD || echo "0000000000000000000000000000000000000000")
    printf '{ version = "%s"; }' "$v" > $out
  '');

  # Need to keep .git around for version string
  srcCleaner = name: type: let baseName = baseNameOf (toString name); in (baseName == ".git" || lib.cleanSourceFilter name type);

  buildGo117Module = buildGoModule.override { go = go_1_17; };
in
buildGo117Module rec {
  pname = "eteu-pdfgen";
  version = if (rev != null) then rev else (versionInfo src).version;

  src = lib.cleanSourceWith { filter = srcCleaner; src = ./.; };

  ldflags = [
    "-X github.com/eteu-technologies/pdfgen/internal/core.Version=${version}"
  ];

  doCheck = true;

  vendorSha256 = "sha256-6z0Ophzh3q9bluXlNQ9YRX/FHEclAioMQrbi6GgBkTo=";
  subPackages = [ "cmd/pdfgen" ];
}
