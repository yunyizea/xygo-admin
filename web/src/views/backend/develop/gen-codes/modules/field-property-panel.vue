<!-- +----------------------------------------------------------------------
  | XYGo Admin [ Vue3 + GoFrame 企业级中后台管理系统 ]
  +----------------------------------------------------------------------
  | Copyright (c) 2026 大连星韵网络科技有限公司 All rights reserved.
  +----------------------------------------------------------------------
  | Licensed ( https://opensource.org/licenses/MIT )
  +----------------------------------------------------------------------
  | Author: 喜羊羊 <751300685@qq.com>
  +---------------------------------------------------------------------- -->
<!-- 字段属性面板（右栏） -->
<template>
  <div class="property-panel">
    <div v-if="!field" class="property-panel__empty">
      <ArtSvgIcon icon="ri:cursor-line" class="text-3xl text-gray-300 mb-2" />
      <p class="text-xs text-gray-400">点击中间区域的字段查看属性</p>
    </div>

    <template v-else>
      <!-- 通用属性 -->
      <div class="prop-section">
        <div class="prop-section__title">通用</div>
        <div class="prop-row">
          <label class="prop-label">设计类型</label>
          <ElSelect v-model="field.designType" size="small" filterable class="prop-input" @change="onDesignTypeChange">
            <ElOption v-for="dt in designTypeList" :key="dt.value" :label="dt.label" :value="dt.value" />
          </ElSelect>
        </div>
        <div class="prop-row">
          <label class="prop-label">字段注释</label>
          <ElInput v-model="field.comment" size="small" class="prop-input" />
        </div>
      </div>

      <!-- 字段属性 -->
      <div class="prop-section">
        <div class="prop-section__title">字段属性</div>
        <div class="prop-row">
          <label class="prop-label">字段名</label>
          <ElInput v-model="field.name" size="small" class="prop-input" :disabled="field.id > 0" @change="onFieldNameChange" />
          <span v-if="field.id > 0" class="text-xs text-gray-400 mt-0.5">已有字段不可改名</span>
        </div>
        <div class="prop-row">
          <label class="prop-label">数据库类型</label>
          <ElInput v-model="field.dbType" size="small" class="prop-input" />
        </div>
        <div class="prop-row">
          <label class="prop-label">Go 类型</label>
          <ElInput v-model="field.goType" size="small" class="prop-input" />
        </div>
        <div class="prop-row">
          <label class="prop-label">TS 类型</label>
          <ElInput v-model="field.tsType" size="small" class="prop-input" />
        </div>
        <div class="prop-row">
          <label class="prop-label">列表显示</label>
          <ElSwitch v-model="field.isList" :active-value="1" :inactive-value="0" size="small" />
        </div>
        <div class="prop-row">
          <label class="prop-label">表单编辑</label>
          <ElSwitch v-model="field.isEdit" :active-value="1" :inactive-value="0" size="small" />
        </div>
        <div class="prop-row">
          <label class="prop-label">搜索条件</label>
          <ElSwitch v-model="field.isQuery" :active-value="1" :inactive-value="0" size="small" />
        </div>
        <div class="prop-row">
          <label class="prop-label">必填</label>
          <ElSwitch v-model="field.isRequired" :active-value="1" :inactive-value="0" size="small" />
        </div>
        <div v-if="field.isQuery" class="prop-row">
          <label class="prop-label">查询方式</label>
          <ElSelect v-model="field.queryType" size="small" class="prop-input">
            <ElOption v-for="opt in queryTypes" :key="opt.value" :label="opt.label" :value="opt.value" />
          </ElSelect>
        </div>
        <!-- 字典类型已移除，使用字段注释中的静态选项映射 -->
      </div>

      <!-- 表格属性 -->
      <div v-if="Object.keys(tableProps).length" class="prop-section">
        <div class="prop-section__title">表格属性</div>
        <div v-for="(prop, key) in tableProps" :key="key" class="prop-row">
          <label class="prop-label">{{ prop.label || key }}</label>
          <template v-if="prop.type === 'select'">
            <ElSelect v-model="tablePropValues[key]" size="small" class="prop-input">
              <ElOption v-for="(label, val) in prop.options" :key="val" :label="label" :value="val" />
            </ElSelect>
          </template>
          <template v-else-if="prop.type === 'number'">
            <ElInputNumber v-model="tablePropValues[key]" size="small" controls-position="right" class="prop-input" />
          </template>
          <template v-else-if="prop.type === 'switch'">
            <ElSwitch v-model="tablePropValues[key]" />
          </template>
          <template v-else>
            <ElInput v-model="tablePropValues[key]" size="small" class="prop-input" :placeholder="prop.placeholder" />
          </template>
        </div>
      </div>

      <!-- 表单属性 -->
      <div v-if="Object.keys(formProps).length" class="prop-section">
        <div class="prop-section__title">表单属性</div>
        <div v-for="(prop, key) in formProps" :key="key" class="prop-row">
          <label class="prop-label">{{ prop.label || key }}</label>

          <!-- 关联表选择 -->
          <template v-if="prop.type === 'remoteTableSelect'">
            <ElSelect
              v-model="formPropValues[key]"
              size="small"
              class="prop-input"
              filterable
              :placeholder="prop.placeholder"
              @change="onRemoteTableChange"
            >
              <ElOption
                v-for="t in tableOptions"
                :key="t.tableName"
                :label="`${t.tableName} (${t.tableComment || ''})`"
                :value="t.tableName"
              />
            </ElSelect>
          </template>

          <!-- 关联字段单选 -->
          <template v-else-if="prop.type === 'remoteColumnSelect'">
            <ElSelect
              v-model="formPropValues[key]"
              size="small"
              class="prop-input"
              filterable
              :placeholder="prop.placeholder"
              :disabled="!remoteColumns.length"
            >
              <ElOption
                v-for="col in remoteColumns"
                :key="col.columnName"
                :label="`${col.columnName} (${col.columnComment || col.dataType})`"
                :value="col.columnName"
              />
            </ElSelect>
          </template>

          <!-- 关联字段完整设计器 -->
          <!-- 关联字段配置已移到 Tab 式设计器 -->
          <template v-else-if="prop.type === 'relationFieldsDesigner'">
            <span class="text-xs text-gray-400">请在上方「关联表」Tab 中配置</span>
          </template>

          <!-- 选项编辑器（radio/select/checkbox/switch） -->
          <template v-else-if="prop.type === 'optionsEditor'">
            <div class="options-editor">
              <div v-for="(opt, oi) in dictOptions" :key="oi" class="options-editor__row">
                <ElInput v-model="opt.key" size="small" placeholder="值" style="width:70px" @change="syncDictOptions" />
                <span class="options-editor__sep">=</span>
                <ElInput v-model="opt.label" size="small" placeholder="标签" style="flex:1" @change="syncDictOptions" />
                <ElButton size="small" type="danger" text @click="removeDictOption(oi)">
                  <ElIcon><Delete /></ElIcon>
                </ElButton>
              </div>
              <ElButton size="small" @click="addDictOption" style="width:100%;margin-top:4px">+ 添加选项</ElButton>
            </div>
          </template>

          <!-- 其他类型（保持不变） -->
          <template v-else-if="prop.type === 'selects'">
            <ElSelect v-model="formPropValues[key]" size="small" class="prop-input" multiple filterable>
              <ElOption v-for="(label, val) in prop.options" :key="val" :label="label" :value="val" />
            </ElSelect>
          </template>
          <template v-else-if="prop.type === 'select'">
            <ElSelect v-model="formPropValues[key]" size="small" class="prop-input">
              <ElOption v-for="(label, val) in prop.options" :key="val" :label="label" :value="val" />
            </ElSelect>
          </template>
          <template v-else-if="prop.type === 'textarea'">
            <ElInput v-model="formPropValues[key]" size="small" class="prop-input" type="textarea" :rows="2" :placeholder="prop.placeholder" />
          </template>
          <template v-else-if="prop.type === 'number'">
            <ElInputNumber v-model="formPropValues[key]" size="small" controls-position="right" class="prop-input" />
          </template>
          <template v-else-if="prop.type === 'switch'">
            <ElSwitch v-model="formPropValues[key]" />
          </template>
          <template v-else-if="prop.type !== 'hidden'">
            <ElInput v-model="formPropValues[key]" size="small" class="prop-input" :placeholder="prop.placeholder" />
          </template>
        </div>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
  import { Delete } from '@element-plus/icons-vue'
  import ArtSvgIcon from '@/components/core/base/art-svg-icon/index.vue'
  import { designTypes, snakeToPascal, snakeToCamel, type PropItem } from './field-library'
  import { fetchGenCodesTableSelect, fetchGenCodesColumnList } from '@/api/backend/develop/genCodes'

  // 关联字段配置项
  interface RelationFieldConfig {
    field: string
    label: string          // 中文标签（从注释提取）
    inList: boolean
    inSearch: boolean
    inExport: boolean
    searchType: string     // like | eq | between
    searchComponent: string // input | select | number 等
    listRender: string     // text | tag | image | link 等
  }

  const props = defineProps<{
    field: any | null
    selects: any
  }>()

  const queryTypes = computed(() => props.selects?.queryType || [
    { value: 'eq', label: '精确匹配(=)' },
    { value: 'like', label: '模糊匹配(LIKE)' },
    { value: 'between', label: '区间(BETWEEN)' },
  ])

  // designType 列表（从 selects 或本地字段库获取）
  const designTypeList = computed(() => {
    if (props.selects?.designTypes?.length) return props.selects.designTypes
    return Object.entries(designTypes).map(([value, def]) => ({
      value,
      label: def.name,
    }))
  })

  // ==================== 关联表选择 ====================
  const tableOptions = ref<any[]>([])
  const remoteColumns = ref<any[]>([])
  const relationFieldsConfig = ref<RelationFieldConfig[]>([])

  // 加载数据库表列表
  const loadTableOptions = async () => {
    if (tableOptions.value.length) return
    try {
      const res = await fetchGenCodesTableSelect()
      tableOptions.value = res.list || []
    } catch { /* ignore */ }
  }

  // 关联表变更 -> 加载该表字段
  const onRemoteTableChange = async (tableName: string) => {
    remoteColumns.value = []
    relationFieldsConfig.value = []
    if (!tableName) return
    try {
      const res = await fetchGenCodesColumnList(tableName)
      remoteColumns.value = (res.list || []).map((col: any) => ({
        columnName: col.name || col.columnName,
        columnComment: col.comment || col.columnComment || '',
        dataType: col.dbType || col.dataType || '',
      }))
      // 自动推导显示Label字段（优先 name > title > label > nickname，都没有则留空）
      if (props.field?._formProps) {
        const currentField = props.field._formProps['remote-field']
        const hasField = (n: string) => remoteColumns.value.some((c: any) => c.columnName === n)
        if (!currentField || !hasField(currentField)) {
          const candidates = ['name', 'title', 'label', 'nickname', 'username']
          const found = candidates.find(c => hasField(c))
          props.field._formProps['remote-field'] = found || ''
        }
      }
    } catch { /* ignore */ }
    onRelationConfigChange()
  }

  // 字段名变化时联动更新 GoName / TsName
  const onFieldNameChange = (newName: string) => {
    if (!props.field || !newName) return
    props.field.goName = snakeToPascal(newName)
    props.field.tsName = snakeToCamel(newName)
  }

  // ==================== 选项编辑器（radio/select/checkbox/switch） ====================
  const dictOptions = ref<{ key: string; label: string }[]>([])

  const syncDictOptions = () => {
    if (!props.field) return
    // 确保 _formProps 存在（新导入的字段可能没有）
    if (!props.field._formProps) props.field._formProps = {}
    // 存为 "key1=label1,key2=label2" 格式
    props.field._formProps['dict-options'] = dictOptions.value
      .filter(o => o.key !== '' || o.label !== '')
      .map(o => `${o.key}=${o.label}`)
      .join(',')
  }

  const addDictOption = () => {
    dictOptions.value.push({ key: '', label: '' })
  }

  const removeDictOption = (idx: number) => {
    dictOptions.value.splice(idx, 1)
    syncDictOptions()
  }

  // 关联字段配置变更 -> 序列化为 JSON 写入 _formProps
  const onRelationConfigChange = () => {
    if (props.field?._formProps) {
      props.field._formProps['relation-fields-config'] = JSON.stringify(relationFieldsConfig.value)
      // 同时生成兼容旧逻辑的逗号分隔字段
      props.field._formProps['relation-fields'] = relationFieldsConfig.value.filter(f => f.inList).map(f => f.field).join(',')
      props.field._formProps['relation-search-fields'] = relationFieldsConfig.value.filter(f => f.inSearch).map(f => f.field).join(',')
      props.field._formProps['relation-export-fields'] = relationFieldsConfig.value.filter(f => f.inExport).map(f => f.field).join(',')
    }
  }

  // 从注释解析字典选项（如 "状态:0=成功,1=失败" -> [{key:'0',label:'成功'},{key:'1',label:'失败'}]）
  // 兼容空格分隔写法："状态:1=待审核 2=审核通过" -> 自动转逗号，与后端 parseRadioOptions 对齐
  const parseCommentOptions = (comment: string): { key: string; label: string }[] => {
    if (!comment) return []
    let normalized = comment.replace(/，/g, ',').replace(/：/g, ':').replace(/；/g, ';')
    if (!normalized.includes(':') || !normalized.includes('=')) return []
    const ci = normalized.indexOf(':')
    if (ci < 0 || ci + 1 >= normalized.length) return []
    const before = normalized.substring(0, ci + 1)
    const after = normalized.substring(ci + 1)
    let result = ''
    for (let i = 0; i < after.length; i++) {
      if (after[i] === ' ') {
        let j = i + 1
        while (j < after.length && after[j] === ' ') j++
        if (j < after.length && after[j] !== '=') {
          result += ','
          i = j - 1
          continue
        }
      }
      result += after[i]
    }
    normalized = before + result
    if (!normalized.includes(',')) return []
    const items = normalized.substring(ci + 1).split(',')
    const opts: { key: string; label: string }[] = []
    for (const item of items) {
      const eqIdx = item.trim().indexOf('=')
      if (eqIdx > 0) {
        const key = item.trim().substring(0, eqIdx).trim()
        const label = item.trim().substring(eqIdx + 1).trim()
        if (key && label) opts.push({ key, label })
      }
    }
    return opts
  }

  // 从注释提取 label（取冒号前部分，对齐后端 extractLabel）
  const extractLabelFromComment = (comment: string, fallback: string): string => {
    if (!comment) return fallback
    const idx = comment.search(/[:：]/)
    return idx > 0 ? comment.substring(0, idx) : comment
  }

  // 选择关联字段时自动填充 label
  const onRelFieldSelect = (item: RelationFieldConfig, fieldName: string) => {
    const col = remoteColumns.value.find((c: any) => c.columnName === fieldName)
    if (col) {
      item.label = extractLabelFromComment(col.columnComment, fieldName)
    } else {
      item.label = fieldName
    }
    onRelationConfigChange()
  }

  // 添加关联字段
  const addRelationField = () => {
    relationFieldsConfig.value.push({
      field: '',
      label: '',
      inList: true,
      inSearch: false,
      inExport: true,
      searchType: 'like',
      searchComponent: 'input',
      listRender: 'text',
    })
  }

  // 移除关联字段
  const removeRelationField = (idx: number) => {
    relationFieldsConfig.value.splice(idx, 1)
    onRelationConfigChange()
  }

  // 监听 field 变化，如果是 remoteSelect 类型则加载相关数据
  watch(() => props.field, async (newField) => {
    if (!newField) return

    // 恢复选项编辑器数据
    const dictStr = newField._formProps?.['dict-options'] || ''
    if (dictStr) {
      dictOptions.value = dictStr.split(',').map((item: string) => {
        const eqIdx = item.indexOf('=')
        if (eqIdx <= 0) return { key: item.trim(), label: '' }
        return { key: item.substring(0, eqIdx).trim(), label: item.substring(eqIdx + 1).trim() }
      }).filter((o: any) => o.key || o.label)
    } else {
      // 未手动配置时，尝试从注释中自动解析字典（如 "状态:0=成功,1=失败"）
      dictOptions.value = parseCommentOptions(newField.comment || '')
      // 解析出选项后立即回写到 _formProps，确保提交时 extra 里有 dict-options
      if (dictOptions.value.length > 0) {
        syncDictOptions()
      }
    }

    if (newField.designType === 'remoteSelect' || newField.designType === 'remoteSelects') {
      await loadTableOptions()
      const currentTable = newField._formProps?.['remote-table']
      if (currentTable) {
        await onRemoteTableChange(currentTable)
      }
      // 恢复关联字段配置
      const configStr = newField._formProps?.['relation-fields-config'] || ''
      if (configStr) {
        try {
          relationFieldsConfig.value = JSON.parse(configStr)
        } catch {
          relationFieldsConfig.value = []
        }
      } else {
        // 兼容旧数据：从 relation-fields 逗号分隔字符串恢复
        const rf = newField._formProps?.['relation-fields'] || ''
        if (rf) {
          relationFieldsConfig.value = rf.split(',').map((s: string) => s.trim()).filter(Boolean).map((f: string) => ({
            field: f, label: f, inList: true, inSearch: false, inExport: true,
            searchType: 'like', searchComponent: 'input', listRender: 'text',
          }))
        } else {
          relationFieldsConfig.value = []
        }
      }
    }
  }, { immediate: true })

  // 当前 designType 的表格属性模板
  const tableProps = computed<Record<string, PropItem>>(() => {
    if (!props.field?.designType) return {}
    return designTypes[props.field.designType]?.table || {}
  })

  // 当前 designType 的表单属性模板
  const formProps = computed<Record<string, PropItem>>(() => {
    if (!props.field?.designType) return {}
    return designTypes[props.field.designType]?.form || {}
  })

  // 表格属性值（存在字段的 _tableProps 上）
  const tablePropValues = computed({
    get() {
      if (!props.field) return {}
      if (!props.field._tableProps) props.field._tableProps = {}
      // 初始化默认值
      for (const [key, prop] of Object.entries(tableProps.value)) {
        if (props.field._tableProps[key] === undefined) {
          props.field._tableProps[key] = prop.value
        }
      }
      return props.field._tableProps
    },
    set(val) {
      if (props.field) props.field._tableProps = val
    }
  })

  // 表单属性值（存在字段的 _formProps 上）
  const formPropValues = computed({
    get() {
      if (!props.field) return {}
      if (!props.field._formProps) props.field._formProps = {}
      // 初始化默认值
      for (const [key, prop] of Object.entries(formProps.value)) {
        if (props.field._formProps[key] === undefined) {
          props.field._formProps[key] = prop.value
        }
      }
      return props.field._formProps
    },
    set(val) {
      if (props.field) props.field._formProps = val
    }
  })

  // 切换 designType 时，重置表格/表单属性为新类型的默认值
  const onDesignTypeChange = (newType: string) => {
    if (!props.field) return
    const def = designTypes[newType]
    if (!def) return

    // 重置 _tableProps
    const newTableProps: Record<string, any> = {}
    for (const [key, prop] of Object.entries(def.table)) {
      newTableProps[key] = prop.value
    }
    props.field._tableProps = newTableProps

    // 重置 _formProps（保留已有的 dict-options 和关联配置，防止切换类型时丢失选项）
    const oldFormProps = props.field._formProps || {}
    const preserveKeys = ['dict-options', 'remote-table', 'remote-pk', 'remote-field', 'relation-fields-config', 'relation-fields', 'relation-search-fields', 'relation-export-fields']
    const newFormProps: Record<string, any> = {}
    for (const [key, prop] of Object.entries(def.form)) {
      newFormProps[key] = (preserveKeys.includes(key) && oldFormProps[key]) ? oldFormProps[key] : prop.value
    }
    props.field._formProps = newFormProps

    // 更新 formType 映射
    const formTypeMap: Record<string, string> = {
      pk: 'input', string: 'input', number: 'inputNumber', float: 'inputNumber',
      switch: 'switch', radio: 'radio', checkbox: 'checkbox', select: 'select',
      selects: 'select', textarea: 'textarea', password: 'input',
      datetime: 'datetime', date: 'date', time: 'input', timestamp: 'datetime',
      image: 'imageUpload', images: 'imagesUpload', file: 'fileUpload', files: 'fileUpload',
      editor: 'richEditor', color: 'colorPicker', icon: 'iconSelector',
      city: 'input', remoteSelect: 'remoteSelect', remoteSelects: 'remoteSelect',
      weigh: 'inputNumber',
    }
    props.field.formType = formTypeMap[newType] || 'input'

    // 如果切换到 remoteSelect/remoteSelects，加载表列表
    if (newType === 'remoteSelect' || newType === 'remoteSelects') {
      loadTableOptions()
    }
  }
</script>

<style scoped>
  @reference '@styles/core/tailwind.css';

  .property-panel {
    padding: 0 2px;
  }

  .property-panel__empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 200px;
    color: var(--el-text-color-placeholder);
  }

  .prop-section {
    margin-bottom: 16px;
  }

  .prop-section__title {
    font-size: 12px;
    font-weight: 600;
    color: var(--el-text-color-secondary);
    padding: 6px 0;
    margin-bottom: 6px;
    border-bottom: 1px solid var(--el-border-color-lighter);
    text-transform: uppercase;
    letter-spacing: 1px;
  }

  .prop-row {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 8px;
  }

  .prop-label {
    flex-shrink: 0;
    width: 72px;
    font-size: 12px;
    color: var(--el-text-color-regular);
    text-align: right;
  }

  .prop-input {
    flex: 1;
    min-width: 0;
  }

  /* 选项编辑器 */
  .options-editor {
    width: 100%;
  }

  .options-editor__row {
    display: flex;
    align-items: center;
    gap: 4px;
    margin-bottom: 4px;
  }

  .options-editor__sep {
    font-size: 12px;
    color: var(--el-text-color-secondary);
    flex-shrink: 0;
  }

  /* 关联字段设计器 */
  .relation-designer {
    width: 100%;
  }

  .relation-designer__empty {
    font-size: 12px;
    color: var(--el-text-color-placeholder);
    text-align: center;
    padding: 12px 0;
  }

  .relation-designer__toolbar {
    margin-bottom: 8px;
  }

  .relation-designer__item {
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 6px;
    padding: 8px;
    margin-bottom: 8px;
    background: var(--el-fill-color-lighter);
  }

  .relation-designer__header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 6px;
  }

  .relation-designer__props {
    display: flex;
    gap: 12px;
    margin-bottom: 4px;
    font-size: 12px;
  }

  .relation-designer__props label {
    display: flex;
    align-items: center;
    gap: 2px;
    cursor: pointer;
  }

  .relation-designer__sub {
    display: flex;
    align-items: center;
    gap: 6px;
    margin-top: 4px;
  }

  .relation-designer__sub-label {
    flex-shrink: 0;
    font-size: 11px;
    color: var(--el-text-color-secondary);
    width: 56px;
    text-align: right;
  }
</style>
