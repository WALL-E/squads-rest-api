#!/usr/bin/env bash
set -euo pipefail

API="http://localhost:8090"

echo "🚀 开始测试 Squads REST API"

######################################
# 健康检查
######################################
echo -e "\n===> 健康检查"
curl -s ${API}/health | jq .

######################################
# 创建 Multisig
######################################
echo -e "\n===> 创建 multisig"
MS1=$(curl -s -X POST ${API}/multisigs \
  -H "Content-Type: application/json" \
  -d '{"multisig_address":"0xabc123","name":"Squad A","description":"First Squad"}' | jq -r '.id')

MS2=$(curl -s -X POST ${API}/multisigs \
  -H "Content-Type: application/json" \
  -d '{"multisig_address":"0xdef456","name":"Squad B","description":"Second Squad"}' | jq -r '.id')

echo "创建的 Multisig IDs: $MS1, $MS2"

######################################
# 查询 Multisigs（分页+搜索+排序）
######################################
echo -e "\n===> 查询 multisigs (分页+搜索+排序)"
curl -s "${API}/multisigs?q=Squad&page=1&page_size=10&sort=created_at:desc" | jq .

######################################
# 更新 Multisig
######################################
echo -e "\n===> 更新 multisig $MS1"
curl -s -X PUT ${API}/multisigs/$MS1 \
  -H "Content-Type: application/json" \
  -d '{"multisig_address":"0xabc999","name":"Squad A Updated","description":"Updated Squad"}' | jq .

######################################
# 创建 Vaults
######################################
echo -e "\n===> 创建 vault"
VA1=$(curl -s -X POST ${API}/vaults \
  -H "Content-Type: application/json" \
  -d "{\"vault_address\":\"0xvault1\",\"multisig_address\":\"0xabc999\"}" | jq -r '.id')

VA2=$(curl -s -X POST ${API}/vaults \
  -H "Content-Type: application/json" \
  -d "{\"vault_address\":\"0xvault2\",\"multisig_address\":\"0xdef456\"}" | jq -r '.id')

echo "创建的 Vault IDs: $VA1, $VA2"

######################################
# 创建 Members
######################################
echo -e "\n===> 创建 members"
MB1=$(curl -s -X POST ${API}/members \
  -H "Content-Type: application/json" \
  -d "{\"member_address\":\"0xmem1\",\"name\":\"Alice\",\"multisig_address\":\"0xabc999\"}" | jq -r '.id')

MB2=$(curl -s -X POST ${API}/members \
  -H "Content-Type: application/json" \
  -d "{\"member_address\":\"0xmem2\",\"name\":\"Bob\",\"multisig_address\":\"0xdef456\"}" | jq -r '.id')

echo "创建的 Member IDs: $MB1, $MB2"

######################################
# 查询子资源
######################################
echo -e "\n===> 查询 Multisig $MS1 的 Vaults"
curl -s ${API}/multisigs/$MS1/vaults | jq .

echo -e "\n===> 查询 Multisig $MS1 的 Members"
curl -s ${API}/multisigs/$MS1/members | jq .

######################################
# 删除资源
######################################
echo -e "\n===> 删除 Member $MB1"
curl -s -X DELETE ${API}/members/$MB1 | jq .

echo -e "\n===> 删除 Vault $VA1"
curl -s -X DELETE ${API}/vaults/$VA1 | jq .

echo -e "\n===> 删除 Multisig $MS1"
curl -s -X DELETE ${API}/multisigs/$MS1 | jq .

######################################
# 最终列表检查
######################################
echo -e "\n===> 查询所有 Multisigs"
curl -s ${API}/multisigs | jq .

echo -e "\n===> 查询所有 Vaults"
curl -s ${API}/vaults | jq .

echo -e "\n===> 查询所有 Members"
curl -s ${API}/members | jq .

echo -e "\n✅ 所有测试完成"
