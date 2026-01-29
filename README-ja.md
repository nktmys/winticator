<img src="./src/assets/icons/appicon.png" alt="Winticator logo" title="Winticator" align="left" height="60px" />

# Winticator

[![Go version](https://img.shields.io/github/go-mod/go-version/nktmys/winticator)](https://github.com/nktmys/winticator/blob/main/go.mod)
[![GitHub release](https://img.shields.io/github/release/nktmys/winticator)](https://github.com/nktmys/winticator/releases)
[![Build Status](https://img.shields.io/github/actions/workflow/status/nktmys/winticator/test.yaml?label=go%20test)](https://github.com/nktmys/winticator/actions/workflows/test.yaml)

[English](./README.md) | 日本語

`Winticator` は、デスクトップ環境（Windows / macOS / Linux）向けの独立したオープンソースTOTP認証アプリケーションです。クロスプラットフォームのリファレンス実装として開発されています。

このプロジェクトは、オープンスタンダードに基づいたスタンドアロンのクロスプラットフォームデスクトップアプリケーションの構築・配布方法を示す**技術デモンストレーション**として位置づけられています。

---

## 概要

- クロスプラットフォーム対応デスクトップ認証アプリ（Windows / macOS / Linux）
- オープンスタンダード **TOTP（RFC 6238）** に準拠
- 完全オフライン動作（ネットワーク通信なし）
- シンプルで最小限の実装を目指した設計
- **MITライセンス**のオープンソースソフトウェアとして公開

Winticatorは**技術的な実装とポータビリティ**に重点を置いており、製品としての差別化や商用利用を目的としていません。

---

## プロジェクトの目的

このプロジェクトの主な目的は以下の通りです：

- **クロスプラットフォームデスクトップアプリケーション**構築の実践的なアプローチを示すこと
- デスクトップ環境におけるTOTP認証アプリの**リファレンス実装**を提供すること
- **公開されている標準仕様や情報**を活用した実装知識を共有すること

このプロジェクトは**商用製品を意図したものではなく**、既存のサービスやアプリケーションと競合したり、それらを置き換えることを目的としていません。

---

## 準拠する標準・参考資料

Winticatorは、以下のような公開資料のみに基づいて実装されています：

- RFC 6238：Time-Based One-Time Password Algorithm（時間ベースワンタイムパスワードアルゴリズム）
- TOTPの公開ドキュメントおよび仕様
- オープンソースのリファレンス実装

いかなる組織の独自アルゴリズム、設計、内部仕様もこのプロジェクトでは使用されていません。

---

## 非関連の免責事項

Winticatorは**独立したプロジェクト**です。

- **Google社との提携、承認、関連はありません**
- **特定の企業との提携、承認、関連はありません**
- 記載されている製品名やサービス名は、説明目的でのみ使用されています

---

## 範囲と制限事項

プロジェクトを意図的にシンプルかつ中立に保つため：

- 機能セットは最小限に抑えられています
- 高度なユーザビリティ、エンタープライズ機能、サービス連携は対象外です
- 他の認証アプリケーションとの比較は意図していません

目標は**明確さとシンプルさ**であり、機能の完全性ではありません。

---

## ライセンス

このプロジェクトは**MITライセンス**の下で公開されています。

ライセンス条項に従い、商用目的を含め、本ソフトウェアを自由に使用、改変、配布、組み込むことができます。

詳細は[LICENSE](./LICENSE)ファイルをご覧ください。

---

## 注意事項

このプロジェクトは「現状のまま」提供され、いかなる保証も伴いません。

本番環境やエンタープライズ向けのソリューションをお探しの場合は、公式サポートを提供する確立された製品やサービスをご検討ください。
