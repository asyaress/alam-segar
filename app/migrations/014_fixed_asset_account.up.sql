INSERT INTO chart_of_accounts (owner_id,code,name,category,normal_balance,system_key)
SELECT u.id,'1200','Aset Tetap','asset','debit','fixed_assets'
FROM users u
WHERE u.role='owner'
  AND NOT EXISTS (
      SELECT 1 FROM chart_of_accounts coa
      WHERE coa.owner_id=u.id AND (coa.code='1200' OR coa.system_key='fixed_assets')
  );
