{
  description = "An opinionated rewrite of tree.";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
        version = "0.1.0";
      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "dirtree";
          inherit version;
          src = pkgs.fetchFromGitHub {
            owner = "ohhfishal";
            repo = "dirtree";
            rev = "v${version}";
            hash = "sha256-o7BbcnKAS9DIdQumuqxk3c9jxGhYLAEISBEbD2q9qlg=";
          };
          vendorHash = "sha256-Fx5aHPDPjwe9iomWBJa3yMcuIHx4W2CtHwMg1q62rDI=";
          meta = {
            description = "An opinionated rewrite of tree.";
          };
        };
      }
    );
}
