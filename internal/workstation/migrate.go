package workstation

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ericklucioh/mobdesk/internal/executil"
)

// MigrateUbuntuWorkspace copies projects out of the legacy Ubuntu workspace.
// Ubuntu remains untouched until the user explicitly removes it afterwards.
func (s Service) MigrateUbuntuWorkspace(ctx context.Context) error {
	if _, err := os.Stat(s.Paths.TermuxMigrationDone()); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("verificar migração anterior: %w", err)
	}
	workspace := s.Paths.Workspace()
	if info, err := os.Lstat(workspace); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("workspace do Termux não pode ser link simbólico")
		}
		entries, readErr := os.ReadDir(workspace)
		if readErr != nil {
			return fmt.Errorf("verificar workspace do Termux: %w", readErr)
		}
		if len(entries) > 0 {
			return fmt.Errorf("workspace do Termux já contém projetos; mova-os antes de migrar")
		}
		if err := os.Remove(workspace); err != nil {
			return fmt.Errorf("preparar workspace vazio do Termux: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("verificar workspace do Termux: %w", err)
	}
	temporary, err := os.MkdirTemp(s.Paths.Home, ".mobdesk-migrate-")
	if err != nil {
		return fmt.Errorf("criar workspace temporário: %w", err)
	}
	defer func() { _ = os.RemoveAll(temporary) }()

	command, err := executil.CommandContext(ctx, "proot-distro", "login", "ubuntu", "--", "tar", "-C", s.Paths.UbuntuWorkspace(), "-cf", "-", ".")
	if err != nil {
		return fmt.Errorf("resolver proot-distro: %w", err)
	}
	output, err := command.StdoutPipe()
	if err != nil {
		return fmt.Errorf("abrir exportação do workspace Ubuntu: %w", err)
	}
	command.Stderr = os.Stderr
	if err := command.Start(); err != nil {
		return fmt.Errorf("iniciar exportação do workspace Ubuntu: %w", err)
	}
	extractErr := extractWorkspaceTar(output, temporary)
	waitErr := command.Wait()
	if extractErr != nil {
		return extractErr
	}
	if waitErr != nil {
		return fmt.Errorf("exportar workspace Ubuntu: %w", waitErr)
	}
	if err := os.Chmod(temporary, 0o755); err != nil {
		return fmt.Errorf("definir permissões do workspace migrado: %w", err)
	}
	if err := os.Rename(temporary, workspace); err != nil {
		return fmt.Errorf("ativar workspace migrado: %w", err)
	}
	return s.writePrivateFile(s.Paths.TermuxMigrationDone(), []byte("workspace migrado do Ubuntu\n"))
}

func extractWorkspaceTar(source io.Reader, destination string) error {
	archive := tar.NewReader(source)
	for {
		header, err := archive.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("ler arquivo do workspace Ubuntu: %w", err)
		}
		name := filepath.Clean(header.Name)
		if name == "." {
			continue
		}
		if filepath.IsAbs(name) || name == ".." || strings.HasPrefix(name, ".."+string(filepath.Separator)) {
			return fmt.Errorf("arquivo inválido no workspace Ubuntu: %q", header.Name)
		}
		path := filepath.Join(destination, name)
		relative, err := filepath.Rel(destination, path)
		if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			return fmt.Errorf("arquivo fora do workspace Ubuntu: %q", header.Name)
		}
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(path, 0o755); err != nil {
				return fmt.Errorf("criar diretório migrado %q: %w", name, err)
			}
		case tar.TypeReg, tar.TypeRegA:
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return fmt.Errorf("criar diretório para %q: %w", name, err)
			}
			file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, os.FileMode(header.Mode)&0o777)
			if err != nil {
				return fmt.Errorf("criar arquivo migrado %q: %w", name, err)
			}
			_, copyErr := io.Copy(file, archive)
			closeErr := file.Close()
			if copyErr != nil {
				return fmt.Errorf("copiar arquivo migrado %q: %w", name, copyErr)
			}
			if closeErr != nil {
				return fmt.Errorf("fechar arquivo migrado %q: %w", name, closeErr)
			}
		case tar.TypeSymlink:
			target := filepath.Clean(filepath.Join(filepath.Dir(name), header.Linkname))
			if filepath.IsAbs(header.Linkname) || target == ".." || strings.HasPrefix(target, ".."+string(filepath.Separator)) {
				return fmt.Errorf("link simbólico inválido no workspace Ubuntu: %q", header.Name)
			}
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return fmt.Errorf("criar diretório para link %q: %w", name, err)
			}
			if err := os.Symlink(header.Linkname, path); err != nil {
				return fmt.Errorf("criar link migrado %q: %w", name, err)
			}
		default:
			return fmt.Errorf("tipo de arquivo não suportado no workspace Ubuntu: %q", header.Name)
		}
	}
}
