# CaptMachineTemplate設計と実装

## 背景

ClusterTopologyでは、マシンリソースのテンプレートを定義するためにCaptMachineTemplateが必要です。このドキュメントでは、CaptMachineTemplateの設計判断と実装の詳細について説明します。

## 設計上の考慮点

### 1. 型の重複問題

実装中に遭遇した最初の課題は、CaptMachineTemplateSpecの型の重複でした。

#### 問題
- CaptMachineSetで使用される既存のCaptMachineTemplateSpec
- ClusterTopologyで必要な新しいCaptMachineTemplate

#### 解決策
- 異なる目的のために別々の型を定義
  - CaptMachineTemplateSpec: MachineSetで使用（既存）
  - CaptInfraMachineTemplateSpec: ClusterTopologyで使用（新規）

この分離により：
- 各ユースケースに特化した型定義が可能に
- 既存のコードへの影響を最小限に抑制
- より明確な責務の分離を実現

### 2. テンプレート構造

Cluster APIのパターンに従い、以下の階層構造を採用：

```go
type CaptMachineTemplate struct {
    Spec CaptInfraMachineTemplateSpec
}

type CaptInfraMachineTemplateSpec struct {
    Template CaptInfraMachineTemplateResource
}

type CaptInfraMachineTemplateResource struct {
    Spec CaptInfraMachineTemplateResourceSpec
}
```

この構造により：
- Cluster APIの一貫性のある設計パターンを維持
- テンプレートのバージョニングと更新戦略をサポート
- より柔軟な拡張性を確保

### 3. WorkspaceTemplateとの統合

CaptMachineTemplateは、WorkspaceTemplateを参照してEKSノードグループを作成します。

#### 設計判断
- WorkspaceTemplateRefを必須フィールドとして定義
- コントローラーでWorkspaceTemplateの存在チェックを実装
- テンプレート間の依存関係を明示的に管理

### 4. ノードタイプの抽象化

EKSの異なるノードタイプをサポートするため、NodeType enumを導入：

```go
type NodeType string

const (
    ManagedNodeGroup NodeType = "ManagedNodeGroup"
    Fargate NodeType = "Fargate"
)
```

この抽象化により：
- 将来的なノードタイプの追加が容易
- 型安全性の確保
- 明確なバリデーションルールの適用

## コントローラーの設計

### 1. 責務の明確化

CaptMachineTemplateコントローラーの主な責務：
- テンプレートリソースの監視
- 参照整合性の検証
- イミュータブルな性質の保証

### 2. RBACの考慮

必要最小限の権限を定義：
```go
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=captmachinetemplates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=captmachinetemplates/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=workspacetemplates,verbs=get;list;watch
```

## 実装のベストプラクティス

1. イミュータビリティ
   - テンプレートは作成後に変更不可
   - 新しいバージョンは新しいリソースとして作成

2. バリデーション
   - kubebuilderアノテーションによる宣言的バリデーション
   - コントローラーでの実行時バリデーション

3. エラー処理
   - 明確なエラーメッセージ
   - 適切なログ記録
   - リトライ可能なエラーの識別

## 今後の課題と拡張性

1. ステータス管理
   - テンプレートの使用状況の追跡
   - 依存関係の状態監視

2. バージョニング戦略
   - テンプレートのバージョン管理
   - 更新時の移行パス

3. 高度な機能
   - カスタムバリデーションルール
   - テンプレートの継承メカニズム
   - 条件付き設定

## 結論

CaptMachineTemplateの設計と実装を通じて、以下の点が重要であることが判明しました：

1. 既存のパターンとの整合性
2. 明確な型の分離と責務の定義
3. 将来の拡張性への配慮
4. 適切なバリデーションと制約の設定

これらの設計判断により、ClusterTopologyでのEKSノードグループ管理が効果的に実現できました。
