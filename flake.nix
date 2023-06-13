{
  description = "Systemd userdb";

  inputs.utils.url = "github:numtide/flake-utils";

  outputs = { self, utils, nixpkgs }:
    utils.lib.eachDefaultSystem (system:
      let pkgs = nixpkgs.legacyPackages.${system}; in
      {

        packages.default = pkgs.buildGoModule {
          pname = "go-systemd-userdb";
          version = "0.0.1";
          src = pkgs.lib.cleanSource ./.;
          vendorHash = "sha256-Npmncmq3DXLCD1lVDCgtwCJzwEW1cuWBkTkCF5jaq/w=";
        };

        devShells.default = with pkgs; mkShell {
          nativeBuildInputs = [
            bashInteractive
            go
          ];
        };
      });
}

