<div align="center">
  <img src="https://raw.githubusercontent.com/devicons/devicon/master/icons/go/go-original-wordmark.svg" width="80" />
  <h1>sysmon</h1>
  <p>Dashboard TUI de monitoramento do sistema em tempo real, feito em Go.</p>

  <a href="https://golang.org"><img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat-square&logo=go&logoColor=white" /></a>
  <a href="https://github.com/charmbracelet/bubbletea"><img src="https://img.shields.io/badge/Bubbletea-1.3-FF6B9D?style=flat-square&logo=go&logoColor=white" /></a>
  <a href="https://github.com/charmbracelet/lipgloss"><img src="https://img.shields.io/badge/Lipgloss-1.1-7C3AED?style=flat-square&logo=go&logoColor=white" /></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-22c55e?style=flat-square" /></a>
</div>

<br/>

CPU, memória, disco, rede e processos — tudo num único terminal, atualizado a cada segundo. Inspirado no btop e htop, construído com a stack Charm.

```
╭─────────────────────────────────────────────────────────────────────────────╮
│ SYSMON   DESKTOP-ABC · windows · up 3d 2h 14m                    14:32:07  │
├───────────────────────────────────────┬─────────────────────────────────────┤
│  CPU                                  │  MEMÓRIA                            │
│  Total  ████████████████░░░░  78.4%   │  RAM    ████████████░░░░░  61.2%   │
│  ▁▂▃▄▅▆▇█▇▆▅▄▅▆▇█▅▃▂▁▂▃▅▇█           │         9.8 GB / 16.0 GB            │
│                                       │  ▁▂▃▄▃▄▅▆▅▄▅▆▇█▇▆▅▄▃▄▅▆▇           │
│  Core0  ████████████████░░░░  79.1%   │                                     │
│  Core1  ██████████░░░░░░░░░░  50.3%   │  Swap   ██░░░░░░░░░░░░░░░  12.5%   │
│  Core2  █████████████████░░░  85.7%   │         0.5 GB / 4.0 GB             │
│  Core3  ████████░░░░░░░░░░░░  40.1%   │                                     │
├───────────────────────────────────────┼─────────────────────────────────────┤
│  DISCO                                │  REDE                               │
│  C:\     ████████████░░░░░░  62.3%   │  Ethernet                           │
│          138.2 GB / 221.8 GB          │  ↑ 1.24 MB/s    ↓ 3.81 MB/s        │
│  D:\     ███████░░░░░░░░░░░  43.7%   │                                     │
│          218.5 GB / 500.0 GB          │  ↑ ▁▂▃▄▃▂▁▂   ↓ ▃▅▇█▆▄▂▁          │
├───────────────────────────────────────┴─────────────────────────────────────┤
│  PROCESSOS  [ordenado por CPU]                                               │
│    PID  NOME                            CPU%   MEM%         MEM             │
│   4521  chrome.exe                      12.4    8.3      1.3 GB             │
│   8832  node.exe                         5.2    2.1    340.0 MB             │
│   1204  Code.exe                         3.1    4.7    754.2 MB             │
│    992  svchost.exe                      1.8    0.9    144.0 MB             │
├─────────────────────────────────────────────────────────────────────────────┤
│ [q] Sair   [p] Alternar ordenação (atual: CPU)                              │
╰─────────────────────────────────────────────────────────────────────────────╯
```

---

## Índice

- [Funcionalidades](#funcionalidades)
- [Instalação](#instalação)
- [Uso](#uso)
- [Painéis](#painéis)
- [Atalhos](#atalhos)
- [Cores e thresholds](#cores-e-thresholds)
- [Estrutura](#estrutura)
- [Dependências](#dependências)

---

## Funcionalidades

- **CPU** — uso total com sparkline histórico (60s) + barra por core (até 8 exibidos)
- **Memória** — RAM e Swap com uso absoluto e sparkline histórico
- **Disco** — até 6 partições montadas com barra de uso e tamanhos
- **Rede** — taxas de envio/recebimento por interface + sparklines comparativos
- **Processos** — top 10 por CPU ou Memória, com PID, nome e RSS
- **Layout responsivo** — se adapta ao tamanho do terminal
- **Cores adaptativas** — verde / amarelo / vermelho conforme o uso
- **Atualização contínua** — métricas coletadas a cada segundo, sem bloqueio da UI

---

## Instalação

> **Pré-requisitos** — `go 1.22+`

```bash
git clone https://github.com/user/sysmon
cd sysmon
go build -o sysmon .
```

Ou instalar direto no PATH:

```bash
go install github.com/user/sysmon@latest
```

---

## Uso

```bash
./sysmon
```

O dashboard abre em tela cheia (alternate screen) e começa a coletar métricas imediatamente. Pressione `q` para sair.

> **Windows** — execute como administrador para visualizar métricas de todos os processos.

---

## Painéis

### CPU

Exibe o uso total da CPU com um sparkline dos últimos 60 segundos, seguido de uma barra individual para cada core (até 8; o excedente é indicado como `... +N cores`).

### Memória

Barras de uso para RAM e Swap com os valores absolutos logo abaixo (`usado / total`). O sparkline acompanha o histórico de RAM.

### Disco

Lista as partições montadas com barra de uso percentual e os tamanhos absolutos. Partições sem capacidade reportada são ignoradas.

### Rede

Mostra as taxas de envio (↑) e recebimento (↓) em tempo real para as interfaces ativas. Dois sparklines sincronizados ao final permitem comparar a tendência de cada direção.

### Processos

Tabela com os 10 processos de maior consumo. A ordenação alterna entre CPU% e Mem% com a tecla `p`. O uso de memória exibido é o RSS (Resident Set Size).

---

## Atalhos

| Tecla | Ação |
|-------|------|
| `q` / `Ctrl+C` | Sair |
| `p` | Alternar ordenação dos processos (CPU ↔ Mem) |

---

## Cores e thresholds

A cor das barras muda conforme o percentual de uso:

| Faixa | CPU / Disco | Memória |
|-------|-------------|---------|
| < 70% | 🟢 Verde | 🔵 Azul |
| 70–89% | 🟡 Amarelo | 🟡 Amarelo (a partir de 80%) |
| ≥ 90% | 🔴 Vermelho | 🔴 Vermelho |

---

## Estrutura

```
main.go        entry point — inicia o programa Bubbletea em tela cheia
metrics.go     coleta de dados via gopsutil (CPU, RAM, disco, rede, processos)
model.go       modelo Bubbletea — Init / Update / ring buffer de histórico
view.go        renderização do dashboard e helpers (barras, sparklines, formatação)
styles.go      paleta de cores e estilos lipgloss
```

---

## Dependências

| Pacote | Função |
|--------|--------|
| [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea) | Framework TUI (arquitetura Elm) |
| [charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss) | Estilos, cores e layout |
| [shirou/gopsutil](https://github.com/shirou/gopsutil) | Métricas do sistema (cross-platform) |

---

<div align="center">
  <sub>Built with <a href="https://github.com/charmbracelet/bubbletea">Bubbletea</a> · <a href="https://github.com/charmbracelet/lipgloss">Lipgloss</a> · <a href="https://github.com/shirou/gopsutil">gopsutil</a> · MIT</sub>
</div>
