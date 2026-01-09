{
  description = "Jira CLI";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    jira-src = {
      url = "github:adrian-gierakowski/jira-cli/main";
      flake = false;
    };
  };

  outputs =
    {
      nixpkgs,
      jira-src,
      ...
    }:
    let
      supportedSystems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;
      pkgsFor = system: nixpkgs.legacyPackages.${system};
    in
    {
      packages = forAllSystems (
        system:
        let
          pkgs = pkgsFor system;
        in
        {
          default = pkgs.jira-cli-go.overrideAttrs (oldAttrs: {
            src = jira-src;
            vendorHash = "sha256-8Pkcs3+24qo7YvTUfmp8iqnghFI3xLkLrUg5QZHLe6I=";
          });
        }
      );
    };
}
