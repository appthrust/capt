# Machine概念の導入

## 背景

現在のCAPTの実装では、コンピュートリソース（Fargateプロファイル、マネージドノードグループなど）がControlPlaneリソースの一部として定義されています。これは以下の課題を引き起こしています：

- コンピュートリソースの管理が柔軟性に欠ける
- スケーリングやライフサイクル管理が困難
- Cluster API (CAPI)の標準パターンから逸脱

## 提案

CAPTにMachine概念を導入し、コンピュートリソースを独立したリソースとして管理することを提案します。

### 主な変更点

1. 新しいカスタムリソースの導入
   - CAPTMachineTemplate
   - CAPTMachineDeployment
   - CAPTMachine

2. WorkspaceTemplateの分離
   - ControlPlane用のWorkspaceTemplate
   - Machine用のWorkspaceTemplate

3. Terraform moduleの再構成
   - eks moduleからnode group関連の設定を分離
   - 新しいnode group用のmoduleを作成

## 期待される効果

1. 運用性の向上
   - ノードグループの個別管理が可能に
   - スケーリングの柔軟な制御
   - ライフサイクル管理の改善

2. アーキテクチャの改善
   - 責務の明確な分離
   - CAPIパターンとの整合性
   - 将来的な拡張性の向上
