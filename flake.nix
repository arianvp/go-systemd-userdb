{
  description = "Systemd userdb";

  inputs.utils.url = "github:numtide/flake-utils";

  outputs = { self, utils, nixpkgs }:
    utils.lib.eachDefaultSystem (system:
      let pkgs = nixpkgs.legacyPackages.${system}; in
      {

        defaultPackage = pkgs.buildGoModule {
          pname = "go-systemd-userdb";
          version = "0.0.1";
          src = pkgs.lib.cleanSource ./.;
        };

        devShell = with pkgs; mkShell {
          nativeBuildInputs = [
            bashInteractive
            go
          ];
        };
      });
}

