{
  description = "A Nix-flake-based domain management environment";

  inputs = {
    devshell.url = "github:numtide/devshell";
    flake-parts.url = "https://flakehub.com/f/hercules-ci/flake-parts/0.1.*.tar.gz";
    nixpkgs.url = "https://flakehub.com/f/NixOS/nixpkgs/=0.1.755230.tar.gz";
    systems.url = "github:nix-systems/default";
    treefmt-nix.url = "github:numtide/treefmt-nix";
  };

  outputs =
    inputs:
    inputs.flake-parts.lib.mkFlake { inherit inputs; } {
      systems = import inputs.systems;
      imports = with inputs; [
        devshell.flakeModule
        treefmt-nix.flakeModule
      ];

      perSystem =
        {
          system,
          pkgs,
          ...
        }:
        {
          _module.args.pkgs = import inputs.nixpkgs {
            inherit system;
            overlays = [ (import ./flake-services/overlays { inherit inputs system; }) ];
            config.allowUnfree = true;
          };

          apps.test = {
            program = pkgs.writeShellApplication {
              name = "domain-management-custom-provider-test";
              runtimeInputs = with pkgs; [
                go_1_24
              ];
              text = ''
                go mod tidy
                go test -v ./...
              '';
            };
          };

          apps.lint = {
            program = pkgs.writeShellApplication {
              name = "golangci-lint";
              runtimeInputs = with pkgs; [ golangci-lint ];
              text = "golangci-lint run";
            };
          };

          devshells.default = {
            packages = with pkgs; [
              delve
              gopls
              go_1_24
              go-tools
              terraform
              golangci-lint
              terraform-docs
              terraform-plugin-docs
            ];
          };

          treefmt = {
            programs = {
              gofmt.enable = true;
              nixfmt.enable = true;
              terraform.enable = true;
            };
          };
        };
    };
}
