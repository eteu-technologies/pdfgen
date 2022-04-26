{
  description = "pdfgen";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
  inputs.flake-utils.url = "github:numtide/flake-utils";

  outputs = { self, nixpkgs, flake-utils }:
    let
      supportedSystems = [
        "aarch64-linux"
        "aarch64-darwin"
        "x86_64-linux"
        "x86_64-darwin"
      ];
    in
    flake-utils.lib.eachSystem supportedSystems (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      rec {
        packages.eteu-pdfgen = pkgs.callPackage ./default.nix {
          rev = self.rev or "dirty";
        };

        defaultPackage = packages.eteu-pdfgen;
        defaultApp = {
          type = "app";
          program = "${packages.eteu-pdfgen}/bin/pdfgen";
        };

        devShell = pkgs.mkShell {
          nativeBuildInputs =
            [
              pkgs.go_1_17
              pkgs.golangci-lint
              pkgs.gopls
              pkgs.curl
              pkgs.jq
            ] ++ pkgs.lib.optionals (!pkgs.stdenv.isDarwin) [
              pkgs.chromium
            ];
        };
      });
}
