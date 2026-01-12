{
  description = "Jira CLI development environment";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      {
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            golangci-lint
            delve
          ];

          shellHook = ''
            echo "Welcome to the Jira CLI development environment!"
            echo "Go version: $(go version)"
            echo "Gopls version: $(gopls version | head -n 1)"
          '';
        };
      }
    );
}
