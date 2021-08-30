{ pkgs ? import <nixpkgs> { }, ... }:
pkgs.mkShell {
  nativeBuildInputs = with pkgs; [ go golangci-lint gopls ];
}
