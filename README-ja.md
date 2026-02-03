<img src="./src/assets/icons/appicon.png" alt="Winticator logo" title="Winticator" align="left" height="60px" />

# Winticator

[![Go version](https://img.shields.io/github/go-mod/go-version/nktmys/winticator)](https://github.com/nktmys/winticator/blob/main/go.mod)
[![GitHub release](https://img.shields.io/github/release/nktmys/winticator)](https://github.com/nktmys/winticator/releases/latest)
[![GitHub all releases](https://img.shields.io/github/downloads/nktmys/winticator/total)](https://github.com/nktmys/winticator/releases) 
[![Build Status](https://img.shields.io/github/actions/workflow/status/nktmys/winticator/test.yaml?label=go%20test)](https://github.com/nktmys/winticator/actions/workflows/test.yaml)

[English](./README.md) | 日本語

`Winticator` は、デスクトップ環境（Windows / macOS / Linux）で動作する**TOTP 認証アプリケーション**の、独立したオープンソース実装です。クロスプラットフォーム対応のリファレンスアプリケーションとして実装されています。

本プロジェクトは、公開標準に基づいたスタンドアロン型・クロスプラットフォームデスクトップアプリケーションをどのように構築・配布できるかを示す**技術的デモンストレーション**を目的としています。

---

## 概要

- Windows / macOS / Linux に対応したクロスプラットフォームデスクトップアプリケーション
- 公開標準 **TOTP（RFC 6238）** に基づく実装
- 完全オフライン動作（ネットワーク通信なし）
- シンプルかつ最小限の機能構成
- **MIT License** に基づくオープンソースソフトウェアとして公開

Winticator は**技術的な実装方法や可搬性**に重点を置いており、製品としての差別化や商用利用を目的としていません。

---

## プロジェクトの目的

このプロジェクトの主な目的は以下の通りです：

- **クロスプラットフォームなデスクトップアプリケーション**を構築するための実践的なアプローチを示すこと
- デスクトップ環境におけるTOTP認証アプリケーションの**リファレンス実装**を提供すること
- **公開されている標準仕様や情報**を活用した実装知識を共有すること

このプロジェクトは**商用プロダクトとしての提供を目的としたものではなく**、既存のサービスやアプリケーションと競合・代替する目的としていません。

---

## 準拠規格および参照情報

Winticator は、以下のような**公開資料のみ**に基づいて実装されています：

- RFC 6238：Time-Based One-Time Password Algorithm
- TOTPに関するの公開ドキュメントおよび仕様
- オープンソースのリファレンス実装

特定の企業や組織が保有する**独自アルゴリズム、設計情報、内部仕様等は一切使用していません**。

---

## 非提携に関する免責事項

Winticator は**独立したプロジェクト**です。

- **Google社とは一切の提携・関係・承認を受けていません**
- **特定の企業・団体との提携・関係・承認もありません**
- 記載されている製品名やサービス名は、説明目的でのみ使用されています

---

## 範囲と制限事項

このプロジェクトは、意図的にシンプルかつ中立な構成としています：

- 機能は最小限に限定しています
- 高度なユーザビリティ、企業向け機能、サービス連携等は対象外です
- 他の認証アプリケーションとの比較、評価を目的としていません

このプロジェクトの目的は、**明確さとシンプルさ**であり、機能の網羅性ではありません。

---

## ライセンス

このプロジェクトは**MITライセンス**のもとで公開されています。

ライセンス条項に従う限り、商用利用を含め、自由に使用・改変・再配布・組み込むことができます。

詳細は[LICENSE](./LICENSE)ファイルをご参照ください。

---

## 注意事項

このプロジェクトは**現状のまま（as is）**提供され、いかなる保証も伴いません。

実運用や企業利用を前提としたプロダクションレベルのソリューションをお探しの場合は、公式なサポートを提供している既存の製品・サービスをご検討ください。
