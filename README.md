# ⚡ cryptoterm

> **A blazing-fast, keyboard-driven terminal dashboard for real-time crypto prices — from Bitcoin & stablecoins on Binance to trending memecoins on Solana & Base.**

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8.svg)](https://go.dev)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/fchyoga/cryptoterm/pulls)

---

```
⚡ CRYPTO TERMINAL     [1] Watchlist     [2] Meme Radar     [3] Portfolio     [4] Detail    ● BNC WS  ● DEX  [USD]  07:55:00 UTC
──────────────────────────────────────────────────────────────────────────────────────────────────────────────────
TICKER        SRC    CAT    PRICE          24H CHG     24H HIGH     24H LOW      VOLUME      TREND
▶ BTC          BNC   MAJOR  $77,820.00     + 1.75% ▲   $77,850.00   $76,000.00   $1.04B      ⡠⠤⠒⠊⠉
  ETH          BNC   MAJOR  $2,493.40      + 2.02% ▲   $2,498.68    $2,428.12    $654.80M    ⡠⠤⠒⠊⠉
  SOL          BNC   MAJOR  $106.06        + 6.00% ▲   $106.12      $99.53       $302.75M    ⣀⠤⠒⠊⠉
  BNB          BNC   MAJOR  $753.81        + 3.97% ▲   $759.31      $720.89      $126.86M    ⡠⠤⠒⠊⠉
  USDC         BNC   STABL  $1.00           -0.01% ▼   $1.00        $1.00        $2.80B      ━━━━━━━━━━
  FDUSD        BNC   STABL  $0.9991        + 0.04% ▲   $0.9994      $0.9984      $29.30M     ━━━━━━━━━━
  DOGE         BNC   MEME   $0.0844        + 3.99% ▲   $0.0847      $0.0806      $53.48M     ⡠⠤⠒⠊⠉
  PEPE         BNC   MEME   $0.00000368    + 5.44% ▲   $0.00000370  $0.00000345  $25.53M     ⡠⠤⠒⠊⠉
  WIF          BNC   MEME   $2.14           + 5.30% ▲   $2.25        $1.98        $340.00M    ⡠⠤⠒⠊⠉
  POPCAT (SOL) DEX   MEME   $0.8520        - 3.20% ▼   $0.9100      $0.8300      $95.00M     ⠉⠉⠒⠤⣀
──────────────────────────────────────────────────────────────────────────────────────────────────────────────────
[a] Add  [d] Remove  [/] Filter  [s] Sort  [c] Config  [!] Alert  [r] Refresh  [?] Help  [q] Quit
```

---

## 🌟 Fitur Utama (SLC — Simple, Lovable, Complete)

- ⚡ **Real-Time Live WebSocket**: Data harga dari Binance streaming setiap milidetik langsung dari matching engine.
- 🐶 **DEX & Meme Coin Radar**: Pantau token Solana (Raydium, Pump.fun), Base, Ethereum, dan BSC langsung via DexScreener API.
- 📊 **Unicode Sparklines**: Grafik tren harga mini langsung di dalam baris terminal tanpa dependensi eksternal.
- 💼 **Pelacak Portofolio Spot**: Hubungkan API Key Binance (Read-Only) untuk memantau saldo wallet dan total nilai portofolio dalam USD/IDR.
- 🔔 **Price Alerts**: Pasang target harga koin; terminal akan membunyikan bell audio (`\a`) dan memunculkan toast banner saat harga tersentuh.
- 🇮🇩 **Anti-Blokir (Bebas VPN)**: Menggunakan mirror resmi Binance Vision (`data-stream.binance.vision`), lancar diakses dari jaringan Indonesia (Indihome, Biznet, Telkomsel, Starlink, dll.) tanpa VPN.
- 💾 **Persistensi Otomatis**: Konfigurasi watchlist, alert, dan token kustom tersimpan rapi di `~/.config/cryptoterm/config.json`.

---

## 🚀 Cara Instalasi

Pilih cara instalasi yang paling mudah untuk Anda:

### 1. One-Line Installer (macOS & Linux)
```bash
curl -fsSL https://raw.githubusercontent.com/fchyoga/cryptoterm/main/scripts/install.sh | bash
```

### 2. Via Go Install
Jika Anda sudah memiliki Go terinstall:
```bash
go install github.com/fchyoga/cryptoterm/cmd/cryptoterm@latest
```
*(Pastikan `$GOPATH/bin` ada di dalam `PATH` Anda)*

### 3. Via NPX / NPM (Zero-Install untuk Pengguna Node.js)
```bash
npx cryptoterm
# atau install global
npm install -g cryptoterm
```

### 4. Build Manual dari Source
```bash
git clone https://github.com/fchyoga/cryptoterm.git
cd cryptoterm
make install
```

Setelah terinstall, cukup jalankan:
```bash
cryptoterm
```

---

## 🎮 Kontrol & Shortcut Keyboard

| Tombol | Fungsi |
| :--- | :--- |
| `Tab` / `1 - 4` | Pindah tab: `[1] Watchlist`, `[2] Meme Radar`, `[3] Portfolio`, `[4] Detail View` |
| `↑` / `↓` atau `j` / `k` | Navigasi kursor naik/turun memilih koin |
| `Enter` | Buka tampilan detail statistik koin yang sedang dipilih |
| `a` | **Tambah koin baru** ke Watchlist (simbol Binance atau paste Contract Address DEX) |
| `d` atau `x` | **Hapus koin** yang dipilih dari Watchlist |
| `/` | **Filter / Cari cepat** koin di watchlist yang sedang aktif |
| `s` | **Urutkan (Sort)** daftar berdasarkan: Default, Kenaikan 24h (%), Harga, Volume, atau Nama |
| `c` | **Buka menu Konfigurasi**: Masukkan API Key Binance, ganti mata uang (USD / IDR), toggle suara alert |
| `!` | **Pasang Price Alert** pada koin yang dipilih (bunyi bell terminal saat harga tersentuh) |
| `r` | **Refresh paksa** semua data harga dan saldo portofolio |
| `?` atau `h` | Buka layar bantuan shortcut lengkap |
| `q` / `Ctrl + C` | Keluar dari aplikasi |

---

## 🔑 Konfigurasi API Binance (Opsional)

> 💡 **Monitoring harga publik TIDAK butuh API key apa pun.** Data harga Binance dan DEX langsung berjalan otomatis begitu aplikasi dibuka.

API Key hanya digunakan jika Anda ingin mengaktifkan fitur **Tab [3] Portfolio** untuk memantau saldo spot Binance Anda.

### Langkah Setup API:
1. Buka `cryptoterm`.
2. Tekan tombol **`c`** pada keyboard.
3. Masukkan **Binance API Key** dan **Binance API Secret**.
4. *(Tips Keamanan)*: Buat API Key di dashboard Binance dengan izin **Read-Only**. **JANGAN** aktifkan izin Spot Trading atau Withdrawal.
5. Tekan `Tab` menuju **[ SAVE & APPLY ]** dan tekan `Enter`.
6. Tekan tombol `3` untuk melihat total nilai dan rincian saldo koin spot Anda secara real-time.

Kredensial disimpan secara aman di mesin lokal Anda pada:
```
~/.config/cryptoterm/config.json
```

---

## 💎 Memantau Koin Meme DEX (Solana, Base, Ethereum, BSC)

Token meme yang belum listing di Binance tetap bisa dipantau langsung:
1. Tekan tombol **`a`**.
2. Masukkan simbol atau **paste Smart Contract Address (CA)** token tersebut (contoh: alamat mint token Solana dari Pump.fun / Raydium).
3. Pilih Data Source `DEX (DexScreener)` atau `Auto-Detect`.
4. Tekan `Enter`. Data harga USD, likuiditas, FDV (Market Cap), dan pergerakan 24 jam akan langsung ditampilkan.

---

## 🛠️ Struktur Repositori

```
cryptoterm/
├── cmd/
│   └── cryptoterm/
│       └── main.go              # Entry point aplikasi
├── internal/
│   ├── api/
│   │   ├── binance_ws.go        # Binance WebSocket Client (live stream)
│   │   ├── binance_rest.go      # Binance REST Client & Signed Account Portfolio
│   │   └── dexscreener.go       # DexScreener Client (Meme coin multi-chain)
│   ├── config/
│   │   └── config.go            # Penyimpanan ~/.config/cryptoterm/config.json
│   ├── model/
│   │   └── types.go             # Struct tipe data pasar, koin, portofolio, dan alert
│   └── ui/
│       ├── model.go             # Bubbletea Model & state event loop
│       ├── view.go              # Lipgloss rendering untuk semua tab
│       ├── styles.go            # Tema warna, formatting harga, & badge
│       ├── sparkline.go         # Generator Unicode/Braille sparkline
│       ├── modal_add.go         # Modal tambah koin
│       ├── modal_config.go      # Modal konfigurasi API & preferensi
│       ├── modal_alert.go       # Modal pasang target alert harga
│       └── view_portfolio.go    # Tampilan portofolio saldo spot
├── scripts/
│   └── install.sh               # One-line installer script
├── bin/
│   └── cli.js                   # Node CLI wrapper untuk npx
├── Makefile                     # Build, install, run, cross-compile
├── package.json                 # Konfigurasi NPM
├── go.mod
└── go.sum
```

---

## 🤝 Kontribusi & Dukungan

Kontribusi, *pull request*, dan *issue report* sangat diterima! Jika Anda menyukai project ini, jangan lupa berikan ⭐ **Star** di [GitHub Repository](https://github.com/fchyoga/cryptoterm).

---

## 📜 Lisensi
[MIT License](LICENSE) © 2026 [fchyoga](https://github.com/fchyoga).
