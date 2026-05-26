# Media Archive Converter — 設計ドキュメント

**日付**: 2026-05-26  
**対象プラットフォーム**: Windows (x86_64)  
**実装言語**: Go  
**外部依存**: ffmpeg (libaavif 付きビルド), ffprobe

---

## 1. 概要

Windows 上の指定ソースディレクトリ配下にある画像・動画ファイルを AVIF / AV1 に変換し、ディレクトリツリー構造を保ったまま指定のコピー先ディレクトリへコピーする CLI ツール。  
コピー先はローカルディレクトリ・外部ドライブ・NAS（UNC パス）など OS がアクセスできる任意のパスを指定できる。  
変換完了後は元ファイルと変換中間ファイルを削除する。

---

## 2. 機能要件

| 要件 | 詳細 |
|------|------|
| 対象画像形式 | `.jpg`, `.jpeg`, `.png` → `.avif` |
| 対象動画形式 | `.mp4` → `.mkv` (AV1) |
| 既存形式のパス変更 | `.avif` 画像 / AV1 動画は変換せずコピーのみ |
| スキップ条件 | コピー先の対応ディレクトリに同一ステム名のファイルが存在する場合 |
| ツリー構造維持 | コピー先は `dst / <src からの相対パス>` |
| 元ファイル削除 | コピー成功後に元ファイルを削除 |
| 一時ファイル削除 | コピー成功後に変換中間ファイルを削除 |
| ハードウェアエンコード | AV1: `av1_nvenc` (preset=p7, cq=25, profile=main) |
| エラー時の挙動 | スキップして次のファイルへ継続、ログに記録 |

---

## 3. 全体アーキテクチャ

```
[ソースディレクトリ]
       │
       ▼
  ① ファイル列挙
     filepath.WalkDir で再帰探索
     対象: .jpg/.jpeg/.png/.mp4/.avif/.mkv
       │
       ▼
  ② スキップ判定
     コピー先の対応ディレクトリ内に同一ステム名ファイルが存在すれば SKIP
       │
       ├── 元が AVIF 画像 / AV1 動画
       │       ↓
       │   ③ 直接コピー → コピー先
       │       ↓
       │   元ファイル削除
       │
       └── 変換が必要 (JPEG/PNG/MP4)
               ↓
           ④ 変換 (tmpDir 配下)
               ↓
           ⑤ コピー先へコピー
               ↓
           ⑥ 元ファイル + 一時ファイル削除
       │
       ▼
  ⑦ ログ出力 (stdout + ファイル)
```

---

## 4. 実行モデル

- **画像変換**: goroutine worker pool で並列処理（デフォルト: CPU コア数、`--workers` で上書き可）
- **動画変換**: シリアル実行（`av1_nvenc` が GPU を占有するため並列化しない）
- **コピー先へのコピー**: 変換完了ごとに逐次実行

---

## 5. CLI インターフェース

```
archive-convert.exe [OPTIONS]

Options:
  --src      <path>    変換元ディレクトリ (必須)
  --dst      <path>    コピー先ディレクトリ (ローカル・外部ドライブ・UNC パス等、OS がアクセスできる任意のパス) (必須)
  --workers  <int>     画像変換の並列数 (デフォルト: runtime.NumCPU())
  --log      <file>    ログファイルパス (省略時: stdout のみ)
  --dry-run           実際の変換/コピー/削除を行わずに処理予定を表示
  --help
```

---

## 6. スキップ判定ロジック

```
入力例: src_file = "D:\Screenshots\2024\foo.png"
        dst_base = "E:\Archive"  # ローカル・外部ドライブ・\\NAS\Archive など任意

1. 相対パスを計算: "2024\foo.png"
2. dst のディレクトリ: "E:\Archive\2024\"
3. そのディレクトリ内でステム "foo" を持つファイルを検索
   - "foo.avif" が存在 → SKIP
   - "foo.png"  が存在 → SKIP
   - 存在しない        → 変換・コピーへ進む
```

**AV1 判定** (動画が既に AV1 かどうかの確認):
```bash
ffprobe -v error -select_streams v:0 \
        -show_entries stream=codec_name \
        -of default=noprint_wrappers=1 input.mkv
# 出力: codec_name=av1 → コピーのみ
```

---

## 7. 変換コマンド

### 7.1 画像 → AVIF
```bash
ffmpeg -i input.jpg \
       -c:v libaavif \
       -still-picture 1 \
       -y \
       <tmpDir>/<stem>.avif
```

### 7.2 動画 → MKV (AV1)
```bash
ffmpeg -i input.mp4 \
       -c:v av1_nvenc \
       -preset p7 \
       -cq 25 \
       -profile:v main \
       -c:a copy \
       -y \
       <tmpDir>/<stem>.mkv
```

音声トラックは無変換 (`-c:a copy`) でコピーする。

---

## 8. ファイル操作フロー

```
変換成功時:
  tmpFile → (os.Copy) → コピー先ファイル → (os.Remove) → tmpFile
                                           → (os.Remove) → 元ファイル

変換失敗時:
  (os.Remove) → tmpFile
  元ファイルは保持

コピー失敗時:
  (os.Remove) → tmpFile
  元ファイルは保持
```

一時ディレクトリは実行開始時に `os.MkdirTemp` で作成し、  
プロセス終了時に `defer os.RemoveAll(tmpDir)` でクリーンアップする。

---

## 9. ログ形式

```
[2026-05-26 10:23:01] SKIP   E:\Archive\2024\foo.avif (already exists on dst)
[2026-05-26 10:23:02] SKIP   D:\Screenshots\2024\bar.avif (already AVIF, will copy)
[2026-05-26 10:23:03] COPY   D:\Screenshots\2024\bar.avif → E:\Archive\2024\bar.avif
[2026-05-26 10:23:05] CONV   D:\Screenshots\2024\baz.png → E:\Archive\2024\baz.avif (2.3s)
[2026-05-26 10:23:10] CONV   D:\Screenshots\video.mp4 → E:\Archive\video.mkv (45.2s)
[2026-05-26 10:23:10] ERROR  D:\Screenshots\broken.jpg: ffmpeg exited with code 1: <stderr>
[2026-05-26 10:23:10] DONE   processed=42 skipped=5 errors=1
```

`--log <file>` 指定時は同内容をファイルにも書き込む。

---

## 10. ディレクトリ構成 (実装)

```
WindowsScreenshotArchive/
├── cmd/
│   └── archive-convert/
│       └── main.go          # エントリポイント、CLI パース
├── internal/
│   ├── converter/
│   │   ├── image.go         # AVIF 変換ロジック
│   │   └── video.go         # AV1 変換ロジック
│   ├── walker/
│   │   └── walker.go        # ディレクトリ再帰探索
│   ├── skipper/
│   │   └── skipper.go       # スキップ判定 (コピー先ステム名検索)
│   └── logger/
│       └── logger.go        # ログ出力
├── go.mod
└── docs/
    └── superpowers/
        └── specs/
            └── 2026-05-26-media-archive-converter-design.md
```

---

## 11. 外部依存・前提条件

| 依存 | バージョン要件 | 用途 |
|------|--------------|------|
| ffmpeg | 7.x 以上、libaavif 付きビルド | 画像・動画変換 |
| ffprobe | ffmpeg に同梱 | AV1 コーデック判定 |
| NVIDIA ドライバ | NVENC AV1 対応 (Driver 530+) | av1_nvenc |
| GPU | GeForce RTX 4070 Ti Super | AV1 ハードウェアエンコード |

ffmpeg は PATH に存在するか、`--ffmpeg-path` で明示指定可能とする。

---

## 12. 非機能要件・制約

- 処理はすべてローカル完結（クラウドサービス不使用）
- コピー先はローカルパス・外部ドライブ・UNC パス（NAS 等）を問わず OS がアクセスできる任意のパスを受け付ける。SMB 認証などのマウント操作はツール外。
- AVIF 変換の品質パラメータは現時点で ffmpeg デフォルトを使用（将来的に `--avif-quality` 追加を検討、現スコープ外）
