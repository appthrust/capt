# 実装詳細

## フェーズ1: 基盤の準備

1. 新しいCRDの作成
   - api/v1beta1/captmachinetemplate_types.go
   - api/v1beta1/captmachinedeployment_types.go
   - api/v1beta1/captmachine_types.go

2. コントローラーの実装
   - internal/controller/captmachinetemplate_controller.go
   - internal/controller/captmachinedeployment_controller.go
   - internal/controller/captmachine_controller.go

3. Terraform moduleの分割
   - internal/tf_module/eks_managed_node_group/
   - internal/tf_module/eks_fargate_profile/

## フェーズ2: ControlPlaneからの分離

1. ControlPlaneリソースの修正
   - Fargateプロファイル設定の削除
   - ノードグループ設定の削除
   - Machine関連のリファレンス追加

2. WorkspaceTemplateの分割
   - ControlPlane用のmoduleから不要な部分を削除
   - 新しいMachine用のmoduleを作成

## フェーズ3: 移行サポート

1. 移行ツールの提供
   ```go
   // internal/migration/machine_migration.go
   func MigrateToMachineBasedArchitecture(controlPlane *v1beta1.CAPTControlPlane) (*v1beta1.CAPTMachineTemplate, error) {
     // 既存のControlPlane設定からMachineTemplate設定を生成
   }
   ```

2. 移行ドキュメントの作成
   - 既存クラスタの移行手順
   - 新規クラスタでの推奨設定

## 検証項目

1. 機能検証
   - MachineDeploymentによるスケーリング
   - ノードグループの更新
   - Fargateプロファイルの管理

2. 互換性検証
   - 既存のControlPlane設定との共存
   - アップグレードパス
   - ダウングレードシナリオ

3. パフォーマンス検証
   - リソース作成時間
   - スケーリング応答時間
   - コントローラーのリソース使用量
