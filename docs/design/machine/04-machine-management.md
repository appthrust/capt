# Machine Management in CAPT

## Overview

CAPTのMachine管理は、Cluster API (CAPI)のパターンに従い、3層のリソースによって実現されています：

1. MachineDeployment
2. MachineSet
3. Machine

この階層構造により、効率的なノード管理と柔軟な更新戦略が可能になります。

## Architecture

```
┌─────────────────────┐
│  MachineDeployment  │  更新戦略の管理
├─────────────────────┤  - ローリングアップデート
│     MachineSet      │  レプリカ数の管理
├─────────────────────┤  - スケーリング
│      Machine        │  ノードの管理
└─────────────────────┘  - プロビジョニング
         │
         │ 参照
         ▼
┌─────────────────────┐
│     NodeGroup       │  EKSマネージドノードグループ
└─────────────────────┘
```

## Component Roles

### MachineDeployment
- 更新戦略の管理（RollingUpdate/Recreate）
- MachineSetのライフサイクル管理
- 進行状況の監視
- ロールバック機能の提供

### MachineSet
- 指定されたレプリカ数のMachine管理
- スケーリング操作の制御
- Machineの健全性監視
- テンプレートに基づくMachineの作成

### Machine
- 個々のノードのライフサイクル管理
- NodeGroupへの参照
- WorkspaceTemplateを使用したノードのプロビジョニング
- ノードの状態監視

## Key Features

1. 責務の分離
- 各レベルで明確な責任範囲
- モジュール性の高い設計
- 保守性の向上

2. 柔軟な更新戦略
- ローリングアップデート
- レプリカ数の動的調整
- 段階的なロールアウト

3. NodeGroupとの統合
- 既存のNodeGroupの活用
- EKSアーキテクチャとの整合性
- 効率的なリソース管理

## Benefits

1. 運用性の向上
- 宣言的な設定
- 自動化された更新プロセス
- スケーリングの容易さ

2. 信頼性の向上
- 段階的な更新
- 自動的な健全性チェック
- ロールバック機能

3. 拡張性
- カスタム更新戦略の追加が容易
- 新しいノードタイプへの対応
- 監視機能の拡張

## Next Steps

詳細な情報については、以下のドキュメントを参照してください：

- [Machine Management Design Details](05-machine-management-design.md)
- [Machine Management Implementation Guide](06-machine-management-implementation.md)
