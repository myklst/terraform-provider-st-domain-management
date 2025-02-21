{ inputs, system, ... }:
final: prev: {
  "23.11" = import inputs.nixpkgs-23-11 { system = system; };
  "24.05" = import inputs.nixpkgs-24-05 {
    system = system;
    config.allowUnfree = true;
  };

  # Nixpkgs November 2023 (23.11) has the last instance of go_1_19
  go = final."23.11".go_1_20;
  gopls = final."23.11".gopls;
  delve = final."23.11".delve;
  gotools = final."23.11".gotools;
  golangci-lint = final."23.11".golangci-lint;

  # Pin Terraform to v.1.8.x
  terraform = final."24.05".terraform;
  terraform-docs = final."24.05".terraform-docs;
  terraform-plugin-docs = final."24.05".terraform-plugin-docs;
}
