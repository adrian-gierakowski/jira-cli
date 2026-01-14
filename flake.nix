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
            gofumpt
          ];

          shellHook = ''
            echo "Welcome to the Jira CLI development environment!"
            echo "Go version: $(go version)"
            echo "Gopls version: $(gopls version | head -n 1)"
          '';
        };
        packages.default = pkgs.jira-cli-go.overrideAttrs (oldAttrs: {
          src = pkgs.lib.cleanSource ./.;
          vendorHash = "sha256-8Pkcs3+24qo7YvTUfmp8iqnghFI3xLkLrUg5QZHLe6I=";
        });
      }
    );
}
