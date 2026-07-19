# Mobdesk

[Português](README.md) | [English](README.en.md)

## Leve sua workstation Linux no bolso

O Mobdesk transforma um celular Android em um ambiente pessoal de desenvolvimento. Em vez de depender dos computadores da faculdade, de uma máquina compartilhada ou de várias contas abertas em equipamentos de terceiros, você leva seus projetos, suas ferramentas e seus dados com você.

Com o Mobdesk, o celular funciona como um pequeno servidor de desenvolvimento:

```text
Android
└── Termux — controle do aparelho
    └── Ubuntu persistente — ambiente de trabalho
```

O Ubuntu roda no próprio celular por meio do PRoot-Distro. Você pode trabalhar diretamente no Termux ou conectar outro computador pela rede local usando SSH. Seus arquivos continuam no aparelho sob seu controle.

## Para que serve?

O Mobdesk foi pensado para estudantes, desenvolvedores e profissionais que precisam de um ambiente Linux portátil para:

- estudar programação em C, JavaScript, HTML, React, Java, Go ou Python;
- criar e executar projetos pequenos e médios para estudo e desenvolvimento;
- iniciar servidores locais, como `npm run dev`, e acessá-los pelo navegador;
- usar o celular como uma workstation pessoal durante aulas, viagens ou deslocamentos;
- acessar o mesmo ambiente pelo celular ou por um computador na mesma rede;
- manter código, configurações e sessões sem fazer login em computadores compartilhados.

O objetivo não é substituir uma máquina de produção, uma VM completa ou um desktop gráfico. O Mobdesk é uma workstation móvel, leve e controlada pelo usuário para desenvolvimento, estudo e servidores locais.

## Por que usar o Mobdesk?

### Seu ambiente acompanha você

O ambiente Ubuntu fica persistente no celular. Você não precisa reconstruir sua configuração sempre que trocar de sala, rede ou computador.

### Seus dados permanecem seus

Projetos e configurações ficam no aparelho. Isso reduz a necessidade de deixar GitHub, e-mail, mensageiros ou outras contas pessoais conectadas em computadores compartilhados.

### Um celular, várias possibilidades

Você pode editar código no celular, abrir uma sessão SSH em outro computador e publicar um servidor local para teste no navegador — tudo usando o mesmo ambiente.

### Sem root e sem Docker no celular

O Mobdesk usa Termux e PRoot-Distro. Não exige root, máquina virtual ou Docker real no Android.

## O que está disponível agora?

O MVP atual concentra-se no bootstrap do ambiente:

- instalação do Ubuntu persistente via PRoot-Distro;
- instalação e configuração do OpenSSH no Termux;
- acesso SSH pela porta `8022`;
- abertura da sessão diretamente no Ubuntu;
- detecção do endereço IP local;
- configuração de senha para o acesso SSH;
- operações repetíveis, sem reinstalar o que já existe;
- comandos `setup`, `start` e `stop`.

A TUI completa, a instalação assistida de ferramentas, o gerenciamento de projetos e a central web fazem parte dos próximos estágios do produto. Consulte o [roadmap](../docs/project/ROADMAP.md) para acompanhar essa evolução.

## Instalação para usuários finais

### Requisitos

- um celular Android com arquitetura ARM64, como a maioria dos aparelhos atuais;
- Termux instalado por uma fonte confiável, preferencialmente [F-Droid](https://f-droid.org/packages/com.termux/) ou pelos [releases oficiais](https://github.com/termux/termux-app/releases);
- aproximadamente 1,5 GB livres para o Ubuntu base e espaço adicional para seus projetos;
- uma rede Wi-Fi comum se você quiser acessar o celular por outro computador.

O Mobdesk não requer root. O desempenho depende da memória, temperatura, bateria e das limitações de segundo plano do Android/HyperOS.

### 1. Instale o Mobdesk

Abra o Termux e execute:

```bash
pkg update
pkg upgrade -y
pkg install -y golang git
go install github.com/ericklucioh/mobdesk/cmd/mobdesk@latest
```

### 2. Configure o ambiente

Na primeira execução, use o binário instalado pelo Go:

```bash
~/go/bin/mobdesk setup
```

O setup instala os componentes necessários no Termux, baixa o Ubuntu, cria o workspace persistente e solicita a senha usada no acesso SSH. Ao final, o comando `mobdesk` fica disponível globalmente.

### 3. Inicie sua workstation

```bash
mobdesk start
```

O Mobdesk inicia o SSH na porta `8022`, mantém o aparelho acordado durante o uso e abre uma sessão Ubuntu no próprio Termux.

Para acessar a workstation a partir de outro computador conectado à mesma rede, use o comando SSH exibido pelo Mobdesk, por exemplo:

```bash
ssh -p 8022 android@192.168.1.50
```

Substitua o endereço pelo IP mostrado no seu aparelho e informe a senha configurada no setup. A conexão SSH será direcionada diretamente para o Ubuntu.

### 4. Pare quando terminar

Para sair apenas da sessão Ubuntu, execute:

```bash
exit
```

Para desligar o servidor SSH:

```bash
mobdesk stop
```

## Entenda o fluxo

```text
mobdesk setup
    ↓
Termux + PRoot-Distro + Ubuntu persistente
    ↓
mobdesk start
    ↓
SSH :8022 → sessão Ubuntu
    ↓
projetos, editores e servidores locais
```

O Termux é o host de controle. O Ubuntu é o ambiente de desenvolvimento. O PRoot melhora a compatibilidade com ferramentas Linux, mas não cria um kernel separado nem oferece o isolamento de uma VM ou de um container real.

## Limites importantes

O Mobdesk é adequado para estudo, desenvolvimento e servidores leves. Ele não foi projetado para:

- cargas pesadas de produção;
- testes de carga ou performance em escala;
- Docker real, systemd ou uma VM Linux completa;
- desktop gráfico completo com aceleração garantida;
- acesso privilegiado a dispositivos ou módulos do kernel.

O Android pode suspender ou encerrar o Termux. Para uma experiência mais estável, permita que o Termux seja executado em segundo plano nas configurações de bateria do aparelho.

## Segurança

Use SSH apenas em redes confiáveis. Não exponha a porta `8022` diretamente na internet. Para acesso remoto fora da rede local, prefira uma rede privada como Tailscale ou um túnel SSH. Faça backups dos projetos importantes fora do celular.

## Documentação

- [Missão do produto](../docs/project/MISSAO.md)
- [MVP atual](../docs/project/MVP.md)
- [Roadmap](../docs/project/ROADMAP.md)
- [Arquitetura e limites](../docs/project/ARQUITETURA.md)
- [Contribuição](../docs/CONTRIBUINDO.md)

## Licença

O Mobdesk é distribuído sob a [licença MIT](../LICENSE).
