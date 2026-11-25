# NOTE

記錄正在進行及最近完成的開發事項，供進度掌握與接續開發之用。隨開發進程即時滾動更新，主動淘汰過時資訊，不作永久流水帳；歷史由 Git 保存。未來規劃見 [ROADMAP.md](ROADMAP.md)，維護方式見 [Development workflow](docs/development-workflow.md)。

## Recently Completed

- Response timeout warning（`62e43a6`）：兩端相關測試通過；初版正式服務已重啟，基本功能驗測正常。未刻意觸發上游逾時，正式預警時序尚未實機驗證。詳見 [預警協定](docs/response-timeout-warning.md)。
