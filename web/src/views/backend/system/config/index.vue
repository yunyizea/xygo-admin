<!-- +----------------------------------------------------------------------
  | XYGo Admin [ Vue3 + GoFrame 企业级中后台管理系统 ]
  +----------------------------------------------------------------------
  | Copyright (c) 2026 大连星韵网络科技有限公司 All rights reserved.
  +----------------------------------------------------------------------
  | Licensed ( https://opensource.org/licenses/MIT )
  +----------------------------------------------------------------------
  | Author: 喜羊羊 <751300685@qq.com>
  +---------------------------------------------------------------------- -->
<!-- 系统配置页面 -->
<template>
  <div class="page-content">
    <div class="config-container">
      <!-- 左侧分组列表 -->
      <div class="group-sidebar">
        <div class="group-header">
          <span>配置分组</span>
          <ElButton type="primary" size="small" @click="showAddGroupDialog">添加分组</ElButton>
        </div>
        <div class="group-list">
          <div
            v-for="group in configGroups"
            :key="group.key"
            class="group-item"
            :class="{ active: selectedGroup === group.key }"
          >
            <div class="group-content" @click="selectGroup(group.key)">
              <div class="group-icon">
                <ArtSvgIcon :icon="group.icon" />
              </div>
              <div class="group-info">
                <div class="group-title">{{ group.title }}</div>
                <div class="group-desc">{{ group.description }}</div>
              </div>
            </div>
            <ElDropdown trigger="click" @command="(cmd) => handleGroupAction(cmd, group)">
              <div class="group-menu-btn" @click.stop>
                <ArtSvgIcon icon="ri:more-fill" :size="16" />
              </div>
              <template #dropdown>
                <ElDropdownMenu>
                  <ElDropdownItem command="edit">
                    <ArtSvgIcon icon="ri:edit-line" :size="14" style="margin-right: 6px;" />
                    编辑
                  </ElDropdownItem>
                  <ElDropdownItem command="delete" divided>
                    <ArtSvgIcon icon="ri:delete-bin-line" :size="14" style="margin-right: 6px;" />
                    删除
                  </ElDropdownItem>
                </ElDropdownMenu>
              </template>
            </ElDropdown>
          </div>
        </div>
      </div>

      <!-- 右侧配置项 -->
      <div class="config-main">
        <div class="config-header">
          <div>
            <h3>{{ currentGroupTitle }}</h3>
            <p class="config-desc">{{ currentGroupDesc }}</p>
          </div>
          <ElButton type="primary" size="small" @click="showAddConfigDialog">添加配置</ElButton>
        </div>

        <div class="config-content" v-loading="loading">
          <!-- 动态配置项 -->
          <div v-if="groupSchemaItems.length > 0 && Object.keys(configFormData).length > 0" class="config-section">

            <!-- OSS 分组：按子 tab 分类显示 -->
            <template v-if="hasSubTabs">
              <ElTabs v-model="activeSubTab" class="config-sub-tabs">
                <ElTabPane
                  v-for="tab in subTabs"
                  :key="tab.key"
                  :label="tab.label"
                  :name="tab.key"
                />
              </ElTabs>
            </template>

            <!-- 表单（tab 模式下只显示当前 tab 的 items） -->
            <ElForm :model="configFormData" :label-width="hasSubTabs ? '160px' : '140px'" class="config-form">
              <ElFormItem 
                v-for="schema in displaySchemaItems" 
                :key="schema.key" 
                :label="schema.name"
                :required="isRequired(schema)"
                class="config-form-item"
              >
                <div class="config-item-wrapper">
                  <div class="config-item-content">
                    <!-- 1. 文本输入框 (text/string) -->
                    <ElInput 
                      v-if="schema.type === 'text' || schema.type === 'string'"
                      v-model="configFormData[schema.key]" 
                      :placeholder="schema.remark || `请输入${schema.name}`"
                      clearable
                    />
                    
                    <!-- 2. 密码输入框 (password) -->
                    <ElInput 
                      v-else-if="schema.type === 'password'"
                      v-model="configFormData[schema.key]" 
                      type="password"
                      show-password
                      :placeholder="schema.remark || `请输入${schema.name}`"
                      clearable
                    />
                    
                    <!-- 3. 多行文本 (textarea) -->
                    <ElInput 
                      v-else-if="schema.type === 'textarea'"
                      v-model="configFormData[schema.key]" 
                      type="textarea"
                      :rows="schema.options?.rows || 3"
                      :placeholder="schema.remark || `请输入${schema.name}`"
                    />
                    
                    <!-- 4. 数字输入 (number) -->
                    <ElInputNumber 
                      v-else-if="schema.type === 'number'"
                      v-model.number="configFormData[schema.key]" 
                      :min="schema.options?.min"
                      :max="schema.options?.max"
                      :step="schema.options?.step || 1"
                      :controls-position="'right'"
                      style="width: 100%;"
                    />
                    
                    <!-- 5. 开关 (switch) -->
                    <ElSwitch 
                      v-else-if="schema.type === 'switch'"
                      v-model="configFormData[schema.key]"
                      :active-value="schema.options?.activeValue || '1'"
                      :inactive-value="schema.options?.inactiveValue || '0'"
                      :active-text="schema.options?.activeText || ''"
                      :inactive-text="schema.options?.inactiveText || ''"
                    />
                    
                    <!-- 6. 下拉单选 (select) -->
                    <ElSelect 
                      v-else-if="schema.type === 'select'"
                      v-model="configFormData[schema.key]"
                      :placeholder="schema.remark || `请选择${schema.name}`"
                      clearable
                      filterable
                      style="width: 100%;"
                    >
                      <ElOption 
                        v-for="opt in schema.options?.options || []"
                        :key="opt.value"
                        :label="opt.label"
                        :value="opt.value"
                      />
                    </ElSelect>
                    
                    <!-- 7. 下拉多选 (selects) -->
                    <ElSelect 
                      v-else-if="schema.type === 'selects'"
                      v-model="configFormData[schema.key]"
                      :placeholder="schema.remark || `请选择${schema.name}`"
                      multiple
                      clearable
                      filterable
                      style="width: 100%;"
                    >
                      <ElOption 
                        v-for="opt in schema.options?.options || []"
                        :key="opt.value"
                        :label="opt.label"
                        :value="opt.value"
                      />
                    </ElSelect>
                    
                    <!-- 8. 单选框组 (radio) -->
                    <ElRadioGroup 
                      v-else-if="schema.type === 'radio'"
                      v-model="configFormData[schema.key]"
                    >
                      <ElRadio 
                        v-for="opt in schema.options?.options || []"
                        :key="opt.value"
                        :label="opt.value"
                      >
                        {{ opt.label }}
                      </ElRadio>
                    </ElRadioGroup>
                    
                    <!-- 9. 复选框 (checkbox) -->
                    <ElCheckbox 
                      v-else-if="schema.type === 'checkbox'"
                      v-model="configFormData[schema.key]"
                      :true-label="'1'"
                      :false-label="'0'"
                    >
                      {{ schema.options?.label || schema.name }}
                    </ElCheckbox>
                    
                    <!-- 10. 日期时间选择器 (datetime) -->
                    <ElDatePicker
                      v-else-if="schema.type === 'datetime'"
                      v-model="configFormData[schema.key]"
                      type="datetime"
                      :placeholder="schema.remark || `请选择${schema.name}`"
                      value-format="YYYY-MM-DD HH:mm:ss"
                      style="width: 100%;"
                    />
                    
                    <!-- 11. 日期选择器 (date) -->
                    <ElDatePicker
                      v-else-if="schema.type === 'date'"
                      v-model="configFormData[schema.key]"
                      type="date"
                      :placeholder="schema.remark || `请选择${schema.name}`"
                      value-format="YYYY-MM-DD"
                      style="width: 100%;"
                    />
                    
                    <!-- 12. 年份选择器 (year) -->
                    <ElDatePicker
                      v-else-if="schema.type === 'year'"
                      v-model="configFormData[schema.key]"
                      type="year"
                      :placeholder="schema.remark || `请选择${schema.name}`"
                      value-format="YYYY"
                      style="width: 100%;"
                    />
                    
                    <!-- 13. 时间选择器 (time) -->
                    <ElTimePicker
                      v-else-if="schema.type === 'time'"
                      v-model="configFormData[schema.key]"
                      :placeholder="schema.remark || `请选择${schema.name}`"
                      value-format="HH:mm:ss"
                      style="width: 100%;"
                    />
                    
                    <!-- 14. 颜色选择器 (color) -->
                    <ArtColorPicker
                      v-else-if="schema.type === 'color'"
                      v-model="configFormData[schema.key]"
                      :placeholder="schema.remark || `请选择${schema.name}`"
                    />
                    
                    <!-- 15. 图标选择器 (icon) -->
                    <ArtIconSelector
                      v-else-if="schema.type === 'icon'"
                      v-model="configFormData[schema.key]"
                    />
                    
                    <!-- 16. 单图上传 (image/upload) -->
                    <ArtImageUpload
                      v-else-if="schema.type === 'image' || schema.type === 'upload'"
                      v-model="configFormData[schema.key]"
                      :max-size="schema.options?.maxSize || 5"
                    />
                    
                    <!-- 17. 多图上传 (images) -->
                    <ArtImageUpload
                      v-else-if="schema.type === 'images'"
                      v-model="configFormData[schema.key]"
                      multiple
                      :limit="schema.options?.limit || 9"
                      :max-size="schema.options?.maxSize || 5"
                    />
                    
                    <!-- 18. 单文件上传 (file) -->
                    <ArtFileUpload
                      v-else-if="schema.type === 'file'"
                      v-model="configFormData[schema.key]"
                      :accept="schema.options?.accept"
                      :max-size="schema.options?.maxSize || 10"
                    />
                    
                    <!-- 19. 多文件上传 (files) -->
                    <ArtFileUpload
                      v-else-if="schema.type === 'files'"
                      v-model="configFormData[schema.key]"
                      multiple
                      :limit="schema.options?.limit || 10"
                      :accept="schema.options?.accept"
                      :max-size="schema.options?.maxSize || 10"
                    />
                    
                    <!-- 20. 富文本编辑器 (editor) -->
                    <ArtWangEditor
                      v-else-if="schema.type === 'editor'"
                      v-model="configFormData[schema.key]"
                      :placeholder="schema.remark || `请输入${schema.name}`"
                      :height="schema.options?.height || '200px'"
                    />
                    
                    <!-- 21. JSON/Object 编辑器 (json/object) -->
                    <ElInput 
                      v-else-if="schema.type === 'json' || schema.type === 'object'"
                      v-model="configFormData[schema.key]" 
                      type="textarea"
                      :rows="6"
                      :placeholder="schema.remark || '请输入JSON格式数据'"
                    />
                    
                    <!-- 22. 数组编辑器 (array) -->
                    <ArtArrayEditor
                      v-else-if="schema.type === 'array' && schema.options?.fields && Array.isArray(configFormData[schema.key])"
                      v-model="configFormData[schema.key]"
                      :fields="schema.options.fields"
                      :show-index="schema.options?.showIndex !== false"
                      :sortable="schema.options?.sortable !== false"
                    />
                    <!-- 22.1 简单数组（无字段配置时降级为文本框） -->
                    <ElInput 
                      v-else-if="schema.type === 'array'"
                      v-model="configFormData[schema.key]" 
                      type="textarea"
                      :rows="4"
                      :placeholder="schema.remark || '请输入数组，多个值用逗号分隔'"
                    />
                    
                    <!-- 23. 远程单选 (remoteSelect) -->
                    <ElSelect 
                      v-else-if="schema.type === 'remoteSelect'"
                      v-model="configFormData[schema.key]"
                      :placeholder="schema.remark || `请选择${schema.name}`"
                      clearable
                      filterable
                      remote
                      :remote-method="(query: string) => handleRemoteSearch(query, schema)"
                      style="width: 100%;"
                    >
                      <ElOption 
                        v-for="opt in remoteOptions[schema.key] || []"
                        :key="opt.value"
                        :label="opt.label"
                        :value="opt.value"
                      />
                    </ElSelect>
                    
                    <!-- 24. 远程多选 (remoteSelects) -->
                    <ElSelect 
                      v-else-if="schema.type === 'remoteSelects'"
                      v-model="configFormData[schema.key]"
                      :placeholder="schema.remark || `请选择${schema.name}`"
                      multiple
                      clearable
                      filterable
                      remote
                      :remote-method="(query: string) => handleRemoteSearch(query, schema)"
                      style="width: 100%;"
                    >
                      <ElOption 
                        v-for="opt in remoteOptions[schema.key] || []"
                        :key="opt.value"
                        :label="opt.label"
                        :value="opt.value"
                      />
                    </ElSelect>
                    
                    <!-- 25. 城市选择器 (city) -->
                    <ElCascader
                      v-else-if="schema.type === 'city'"
                      v-model="configFormData[schema.key]"
                      :options="cityOptions"
                      :placeholder="schema.remark || `请选择${schema.name}`"
                      clearable
                      filterable
                      style="width: 100%;"
                    />
                    
                    <!-- 默认文本输入 -->
                    <ElInput 
                      v-else
                      v-model="configFormData[schema.key]" 
                      :placeholder="schema.remark || `请输入${schema.name}`"
                      clearable
                    />
                  </div>

                  <!-- 右侧三点菜单 -->
                  <ElDropdown trigger="click" @command="(cmd) => handleConfigAction(cmd, schema)">
                    <div class="config-menu-btn" @click.stop>
                      <ArtSvgIcon icon="ri:more-fill" :size="16" />
                    </div>
                    <template #dropdown>
                      <ElDropdownMenu>
                        <ElDropdownItem command="copy">
                          <ArtSvgIcon icon="ri:file-copy-line" :size="14" style="margin-right: 6px;" />
                          复制变量名 ({{ schema.key }})
                        </ElDropdownItem>
                        <ElDropdownItem 
                          command="delete" 
                          divided
                          :disabled="schema.allowDel === 0"
                        >
                          <ArtSvgIcon icon="ri:delete-bin-line" :size="14" style="margin-right: 6px;" />
                          <span>删除配置</span>
                          <span v-if="schema.allowDel === 0" class="menu-item-hint">（核心配置）</span>
                        </ElDropdownItem>
                      </ElDropdownMenu>
                    </template>
                  </ElDropdown>
                </div>
                
                <!-- 备注提示 -->
                <div v-if="schema.remark" class="form-item-hint">
                  {{ schema.remark }}
                </div>
              </ElFormItem>
            </ElForm>
          </div>

          <!-- 空状态 -->
          <div v-else class="config-section">
            <ElEmpty description="该分组暂无配置项">
              <ElButton type="primary" @click="showAddConfigDialog">
                添加配置项
              </ElButton>
            </ElEmpty>
          </div>
        </div>
        
        <!-- 底部操作按钮 -->
        <div class="config-footer">
          <ElButton type="primary" @click="handleSave" :loading="loading">
            保存配置
          </ElButton>
        </div>
      </div>
    </div>

    <!-- 添加/编辑分组弹窗 -->
    <ElDialog
      v-model="addGroupVisible"
      :title="isEditMode ? '编辑配置分组' : '添加配置分组'"
      width="500px"
      @close="resetGroupForm"
    >
      <ElForm :model="groupForm" label-width="100px">
        <ElFormItem label="分组键名">
          <ElInput v-model="groupForm.key" placeholder="如：wechat" :disabled="isEditMode" />
          <div v-if="isEditMode" class="form-hint">编辑时不可修改键名</div>
        </ElFormItem>
        <ElFormItem label="分组名称">
          <ElInput v-model="groupForm.name" placeholder="如：微信配置" />
        </ElFormItem>
        <ElFormItem label="图标">
          <ArtIconSelector v-model="groupForm.icon" />
        </ElFormItem>
        <ElFormItem label="分组描述">
          <ElInput v-model="groupForm.description" type="textarea" :rows="2" placeholder="如：微信小程序/公众号参数配置" />
        </ElFormItem>
      </ElForm>
      <template #footer>
        <ElButton @click="addGroupVisible = false">取消</ElButton>
        <ElButton type="primary" @click="confirmAddGroup">
          {{ isEditMode ? '保存' : '确定' }}
        </ElButton>
      </template>
    </ElDialog>

    <!-- 添加配置弹窗（新组件） -->
    <AddConfigItemDialog
      :visible="addConfigVisible"
      :group-options="configGroups"
      :default-group="selectedGroup"
      :lock-group="true"
      @update:visible="val => (addConfigVisible = val)"
      @confirm="handleAddConfigConfirm"
    />
  </div>
</template>

<script setup lang="ts">
import { 
  getConfigGroupListApi, 
  saveConfigGroupApi, 
  deleteConfigGroupApi, 
  getConfigSchemaApi,
  getConfigListApi,
  saveConfigApi,
  createConfigItemApi,
  deleteConfigItemApi,
  type ConfigGroupItem,
  type ConfigItem,
  type ConfigSchemaItem
} from '@/api/backend/system/config'
import { useClipboard } from '@vueuse/core'
import AddConfigItemDialog from './components/AddConfigItemDialog.vue'

const { copy, copied } = useClipboard()

defineOptions({ name: 'SystemConfig' })

// 当前选中分组
const selectedGroup = ref('basics')
const loading = ref(false)
const addGroupVisible = ref(false)
const addConfigVisible = ref(false)
const isEditMode = ref(false)
const editingGroupKey = ref('')

// 配置分组列表
const configGroups = ref<Array<{
  key: string
  title: string
  description: string
  icon: string
  sort?: number
}>>([])

// 当前分组的配置项
const currentConfigItems = ref<ConfigItem[]>([])
// 使用 any 避免多类型（数组、对象、字符串、数值）赋值报错
const configFormData = ref<Record<string, any>>({})

// 所有配置项的 Schema 信息
const configSchemaMap = ref<Record<string, ConfigSchemaItem>>({})
const groupSchemaItems = computed(() => {
  return Object.values(configSchemaMap.value).filter(
    (item: ConfigSchemaItem) => item.group === selectedGroup.value
  ).sort((a, b) => a.sort - b.sort)
})

// 解析 rules，判断是否必填
const parseRules = (rules: any) => {
  if (!rules) return []
  try {
    if (typeof rules === 'string') {
      const parsed = JSON.parse(rules)
      return Array.isArray(parsed) ? parsed : [parsed]
    }
    if (Array.isArray(rules)) return rules
    if (typeof rules === 'object') return [rules]
  } catch (e) {
    console.warn('rules parse failed', rules, e)
  }
  return []
}

const isRequired = (schema: any) => {
  const rules = parseRules(schema.rules)
  return rules.some((r: any) => r && r.required === true)
}

// ==================== 子 Tab 分组（用于 OSS 等有多个子分类的分组）====================
const activeSubTab = ref('general')

// 子 Tab 定义（按 key 前缀分组）
const subTabDefs: Record<string, Array<{ key: string; label: string; prefixes: string[] }>> = {
  oss: [
    { key: 'general', label: '通用配置', prefixes: ['oss_driver', 'upload_'] },
    { key: 'aliyun', label: '阿里云 OSS', prefixes: ['oss_aliyun_'] },
    { key: 'tencent', label: '腾讯云 COS', prefixes: ['oss_cos_'] },
    { key: 'qiniu', label: '七牛云', prefixes: ['oss_qiniu_'] },
  ],
  sms: [
    { key: 'general', label: '基础配置', prefixes: ['sms_timeout', 'sms_strategy', 'sms_enabled_drivers'] },
    { key: 'aliyun', label: '阿里云短信', prefixes: ['sms_aliyun_'] },
    { key: 'tencent', label: '腾讯云短信', prefixes: ['sms_tencent_'] },
  ],
}

const hasSubTabs = computed(() => !!subTabDefs[selectedGroup.value])

const subTabs = computed(() => subTabDefs[selectedGroup.value] || [])

// 当前实际显示的配置项（有 tab 时按 tab 过滤，否则显示全部）
const displaySchemaItems = computed(() => {
  if (!hasSubTabs.value) return groupSchemaItems.value
  const currentTab = subTabs.value.find(t => t.key === activeSubTab.value)
  if (!currentTab) return groupSchemaItems.value
  return groupSchemaItems.value.filter(item =>
    currentTab.prefixes.some(prefix => item.key.startsWith(prefix))
  )
})

// 切换分组时重置 subTab
watch(selectedGroup, () => {
  activeSubTab.value = subTabs.value.length ? subTabs.value[0].key : 'general'
})

// 远程搜索选项
const remoteOptions = ref<Record<string, Array<{ label: string; value: string }>>>({})

// 城市数据（简化版，实际可以从接口获取）
const cityOptions = ref([
  {
    value: 'beijing',
    label: '北京',
    children: [
      { value: 'chaoyang', label: '朝阳区' },
      { value: 'haidian', label: '海淀区' },
      { value: 'dongcheng', label: '东城区' },
    ]
  },
  {
    value: 'shanghai',
    label: '上海',
    children: [
      { value: 'pudong', label: '浦东新区' },
      { value: 'huangpu', label: '黄浦区' },
      { value: 'xuhui', label: '徐汇区' },
    ]
  },
  {
    value: 'guangzhou',
    label: '广州',
    children: [
      { value: 'tianhe', label: '天河区' },
      { value: 'yuexiu', label: '越秀区' },
      { value: 'haizhu', label: '海珠区' },
    ]
  }
])

// 处理远程搜索
const handleRemoteSearch = async (query: string, schema: any) => {
  if (!query) {
    remoteOptions.value[schema.key] = []
    return
  }
  
  try {
    // TODO: 根据 schema.options?.api 调用远程接口
    // const res = await request.get({ url: schema.options?.api, params: { query } })
    // remoteOptions.value[schema.key] = res.data
    
    // 模拟数据
    remoteOptions.value[schema.key] = [
      { label: `${query} - 选项1`, value: `${query}_1` },
      { label: `${query} - 选项2`, value: `${query}_2` },
    ]
  } catch (error) {
    console.error('远程搜索失败:', error)
  }
}

// 加载配置 Schema
const loadConfigSchema = async () => {
  try {
    const res = await getConfigSchemaApi()
    if (res.list && res.list.length > 0) {
      const schemaMap: Record<string, ConfigSchemaItem> = {}
      res.list.forEach((item: ConfigSchemaItem) => {
        // 确保 allowDel 有默认值
        if (item.allowDel === undefined || item.allowDel === null) {
          item.allowDel = 1
        }
        schemaMap[item.key] = item
      })
      configSchemaMap.value = schemaMap
    }
  } catch (error) {
    console.error('加载配置 Schema 失败:', error)
  }
}

// 加载配置分组列表
const loadConfigGroups = async () => {
  try {
    const res = await getConfigGroupListApi()
    if (res.list && res.list.length > 0) {
      configGroups.value = res.list.map((item: ConfigGroupItem) => ({
        key: item.group,
        title: item.groupName,
        description: item.description || '',
        icon: item.icon || 'ri:settings-3-line',
        sort: item.sort || 0
      }))
      
      // 加载第一个分组的配置
      if (configGroups.value.length > 0) {
        selectedGroup.value = configGroups.value[0].key
        await loadGroupConfig(selectedGroup.value)
      }
    }
  } catch (error) {
    console.error('加载配置分组失败:', error)
    ElMessage.error('加载配置分组失败')
  }
}

// 加载指定分组的配置项
const loadGroupConfig = async (group: string) => {
  try {
    loading.value = true
    const res = await getConfigListApi(group)
    currentConfigItems.value = res.list || []
    
    // 确保 schema 已经加载
    if (Object.keys(configSchemaMap.value).length === 0) {
      await loadConfigSchema()
    }
    
    // 根据 schema 类型转换数据
    const formData: Record<string, any> = {}
    const items = res.items || {}
    
    groupSchemaItems.value.forEach((schema: ConfigSchemaItem) => {
      const rawValue = items[schema.key]
      const value = rawValue !== undefined && rawValue !== null ? rawValue : (schema.value || '')
      
      // 根据类型转换数据，确保所有类型都有合法的初始值
      switch (schema.type) {
        case 'number':
          // 数字类型转换为 Number
          if (value === '' || value === null || value === undefined) {
            formData[schema.key] = null
          } else {
            const num = Number(value)
            formData[schema.key] = isNaN(num) ? null : num
          }
          break
          
        case 'switch':
        case 'checkbox':
          // 开关/复选框保持字符串 "1" 或 "0"
          const strValue = String(value || '0')
          formData[schema.key] = (strValue === '1' || strValue === 'true') ? '1' : '0'
          break
          
        case 'text':
        case 'string':
        case 'password':
        case 'icon':
        case 'color':
        case 'image':
        case 'file':
        case 'datetime':
        case 'date':
        case 'year':
        case 'time':
        case 'select':
        case 'radio':
        case 'remoteSelect':
          // 单值字符串类型，确保是字符串
          formData[schema.key] = value ? String(value) : ''
          break
          
        case 'textarea':
        case 'editor':
          // 文本域和编辑器，确保是字符串（不能是 undefined）
          formData[schema.key] = value !== null && value !== undefined ? String(value) : ''
          break
          
        case 'images':
        case 'files':
          // 多图/多文件，保持字符串
          formData[schema.key] = value ? String(value) : ''
          break
          
        case 'selects':
        case 'remoteSelects':
          // 多选，解析为数组
          if (typeof value === 'string') {
            formData[schema.key] = value ? value.split(',').map((v: string) => v.trim()).filter((v: string) => v) : []
          } else if (Array.isArray(value)) {
            formData[schema.key] = value
          } else {
            formData[schema.key] = []
          }
          break
          
        case 'city':
          // 城市选择器，解析为数组
          try {
            if (typeof value === 'string' && value) {
              formData[schema.key] = JSON.parse(value)
            } else if (Array.isArray(value)) {
              formData[schema.key] = value
            } else {
              formData[schema.key] = []
            }
          } catch {
            formData[schema.key] = []
          }
          break
          
        case 'json':
        case 'object':
          // JSON对象类型
          formData[schema.key] = value ? String(value) : '{}'
          break
          
        case 'array':
          // 数组类型 - 如果有字段配置则解析为数组对象，否则保持字符串
          if (schema.options?.fields) {
            try {
              formData[schema.key] = value ? JSON.parse(String(value)) : []
            } catch {
              formData[schema.key] = []
            }
          } else {
            formData[schema.key] = value ? String(value) : '[]'
          }
          break
          
        default:
          // 其他类型默认为空字符串
          formData[schema.key] = value ? String(value) : ''
      }
    })
    
    configFormData.value = formData
  } catch (error: any) {
    console.error('加载配置项失败:', error)
    // 如果分组没有配置项，显示空状态
    currentConfigItems.value = []
    configFormData.value = {}
  } finally {
    loading.value = false
  }
}

// 页面加载时获取分组列表和配置 Schema
onMounted(async () => {
  await loadConfigSchema()
  await loadConfigGroups()
})

// 当前分组标题和描述
const currentGroupTitle = computed(() => {
  const group = configGroups.value.find((g: any) => g.key === selectedGroup.value)
  return group?.title || ''
})

const currentGroupDesc = computed(() => {
  const group = configGroups.value.find((g: any) => g.key === selectedGroup.value)
  return group?.description || ''
})

// 分组表单
const groupForm = reactive({
  key: '',
  name: '',
  icon: '',
  description: ''
})

// 选择分组
const selectGroup = async (key: string) => {
  selectedGroup.value = key
  await loadGroupConfig(key)
}

// 显示添加分组弹窗
const showAddGroupDialog = () => {
  isEditMode.value = false
  addGroupVisible.value = true
}

// 显示编辑分组弹窗
const showEditGroupDialog = (group: any) => {
  isEditMode.value = true
  editingGroupKey.value = group.key
  groupForm.key = group.key
  groupForm.name = group.title
  groupForm.icon = group.icon
  groupForm.description = group.description
  addGroupVisible.value = true
}

// 处理分组操作（编辑/删除）
const handleGroupAction = (command: string, group: any) => {
  if (command === 'edit') {
    showEditGroupDialog(group)
  } else if (command === 'delete') {
    ElMessageBox.confirm(
      `确定要删除分组"${group.title}"吗？删除后该分组下的所有配置项也将被删除。`,
      '删除确认',
      {
        confirmButtonText: '确定删除',
        cancelButtonText: '取消',
        type: 'warning',
      }
    ).then(async () => {
      try {
        loading.value = true
        await deleteConfigGroupApi(group.key)
        ElMessage.success('删除成功')
        
        // 如果删除的是当前选中的分组，切换到第一个分组
        if (selectedGroup.value === group.key && configGroups.value.length > 0) {
          selectedGroup.value = configGroups.value[0].key
        }
        
        // 重新加载分组列表
        await loadConfigGroups()
      } catch (error: any) {
        ElMessage.error(error.message || '删除失败')
      } finally {
        loading.value = false
      }
    }).catch(() => {
      // 用户取消删除
    })
  }
}

// 显示添加配置弹窗
const showAddConfigDialog = () => {
  addConfigVisible.value = true
}

// 重置分组表单
const resetGroupForm = () => {
  groupForm.key = ''
  groupForm.name = ''
  groupForm.icon = ''
  groupForm.description = ''
  isEditMode.value = false
  editingGroupKey.value = ''
}

// 确认添加/编辑分组
const confirmAddGroup = async () => {
  if (!groupForm.key || !groupForm.name) {
    ElMessage.warning('请填写分组键名和名称')
    return
  }
  
  try {
    loading.value = true
    await saveConfigGroupApi({
      group: groupForm.key,
      groupName: groupForm.name,
      icon: groupForm.icon,
      description: groupForm.description,
      sort: configGroups.value.length * 10, // 自动排序
      isEdit: isEditMode.value
    })
    
    ElMessage.success(isEditMode.value ? '保存成功' : '添加成功')
    addGroupVisible.value = false
    resetGroupForm()
    
    // 重新加载分组列表
    await loadConfigGroups()
  } catch (error: any) {
    ElMessage.error(error.message || '操作失败')
  } finally {
    loading.value = false
  }
}

// 处理新增配置确认
const handleAddConfigConfirm = async (payload: any) => {
  try {
    loading.value = true
    // 查找分组名称
    const groupInfo = configGroups.value.find(g => g.key === payload.group)
    await createConfigItemApi({
      group: payload.group,
      groupName: groupInfo?.title || payload.group,
      name: payload.name,
      key: payload.key,
      type: payload.type,
      value: payload.value || '',
      options: payload.options || '',
      rules: payload.rules || '',
      sort: payload.sort ?? 100,
      remark: payload.remark || '',
    })
    ElMessage.success('配置项创建成功')
    addConfigVisible.value = false
    // 刷新 Schema 和当前分组配置
    await loadConfigSchema()
    await loadGroupConfig(selectedGroup.value)
  } catch (error: any) {
    ElMessage.error(error.message || '创建配置项失败')
  } finally {
    loading.value = false
  }
}

// 处理配置项操作（复制变量名/删除配置）
const handleConfigAction = async (command: string, schema: any) => {
  if (command === 'copy') {
    // 复制变量名（key）
    try {
      await copy(schema.key)
      ElMessage.success(`已复制变量名: ${schema.key}`)
    } catch (error) {
      ElMessage.error('复制失败')
    }
  } else if (command === 'delete') {
    // 检查是否允许删除
    if (schema.allowDel === 0) {
      ElMessage.warning('该配置项不允许删除')
      return
    }
    
    // 确认删除
    ElMessageBox.confirm(
      `确定要删除配置项"${schema.name}"吗？`,
      '删除确认',
      {
        confirmButtonText: '确定删除',
        cancelButtonText: '取消',
        type: 'warning',
      }
    ).then(async () => {
      try {
        loading.value = true
        await deleteConfigItemApi(schema.key)
        ElMessage.success('删除成功')
        
        // 重新加载配置Schema和当前分组配置
        await loadConfigSchema()
        await loadGroupConfig(selectedGroup.value)
      } catch (error: any) {
        ElMessage.error(error.message || '删除失败')
      } finally {
        loading.value = false
      }
    }).catch(() => {
      // 用户取消删除
    })
  }
}

// 保存配置
const handleSave = async () => {
  loading.value = true
  try {
    // 将表单数据转换为接口需要的格式
    const items: ConfigItem[] = groupSchemaItems.value.map(schema => {
      const value = configFormData.value[schema.key]
      let stringValue = ''
      
      // 根据类型转换为字符串
      switch (schema.type) {
        case 'number':
          stringValue = value !== null && value !== undefined ? String(value) : ''
          break
        case 'images':
        case 'files':
          // 多图/多文件，保持逗号分隔的字符串
          stringValue = value || ''
          break
        case 'selects':
        case 'remoteSelects':
          // 多选，转为逗号分隔的字符串或JSON
          stringValue = Array.isArray(value) ? value.join(',') : (value || '')
          break
        case 'city':
          // 城市选择器，转为JSON字符串
          stringValue = Array.isArray(value) ? JSON.stringify(value) : (value || '')
          break
        case 'array':
          // 数组类型 - 如果是数组对象则序列化，否则保持原样
          if (Array.isArray(value)) {
            stringValue = JSON.stringify(value)
          } else {
            stringValue = value || '[]'
          }
          break
        case 'json':
        case 'object':
          // JSON/对象类型，保持字符串
          stringValue = value || '{}'
          break
        default:
          // 其他类型转为字符串
          stringValue = value !== null && value !== undefined ? String(value) : ''
      }
      
      return {
        key: schema.key,
        value: stringValue
      }
    })
    
    await saveConfigApi({
      group: selectedGroup.value,
      items
    })
    
    ElMessage.success('保存成功')
    // 重新加载配置
    await loadGroupConfig(selectedGroup.value)
  } catch (error: any) {
    ElMessage.error(error.message || '保存失败')
    console.error(error)
  } finally {
    loading.value = false
  }
}
</script>

<style scoped lang="scss">
.config-container {
  display: flex;
  gap: 20px;
  height: calc(100vh - 200px);
  max-height: calc(100vh - 200px);
}

.group-sidebar {
  width: 280px;
  height: 100%;
  flex-shrink: 0;
  border: 1px solid var(--art-gray-300);
  border-radius: 8px;
  background: var(--default-box-color);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.group-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid var(--art-gray-300);
  font-weight: 600;
  font-size: 14px;
  color: var(--art-gray-800);
}

.group-list {
  flex: 1;
  padding: 8px;
  overflow-y: auto;
}

.group-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px;
  margin-bottom: 4px;
  border-radius: 6px;
  transition: all 0.2s;
  position: relative;

  &:hover {
    background: var(--art-gray-100);

    .group-menu-btn {
      opacity: 1;
    }
  }

  &.active {
    background: var(--theme-color-alpha-10);
    border-left: 3px solid var(--theme-color);
    padding-left: 9px;

    .group-title {
      color: var(--theme-color);
      font-weight: 600;
    }
  }
}

.group-content {
  display: flex;
  align-items: flex-start;
  flex: 1;
  cursor: pointer;
  min-width: 0;
}

.group-menu-btn {
  flex-shrink: 0;
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s;
  color: var(--art-gray-600);
  opacity: 0;

  &:hover {
    background: var(--art-gray-200);
    color: var(--art-gray-800);
  }
}

.group-icon {
  flex-shrink: 0;
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 12px;
  border-radius: 6px;
  background: var(--art-gray-100);
  color: var(--art-gray-600);
  font-size: 18px;
}

.active .group-icon {
  background: var(--theme-color-alpha-10);
  color: var(--theme-color);
}

.group-info {
  flex: 1;
  min-width: 0;
}

.group-title {
  font-size: 14px;
  font-weight: 500;
  color: var(--art-gray-800);
  margin-bottom: 4px;
  transition: all 0.2s;
}

.group-desc {
  font-size: 12px;
  color: var(--art-gray-500);
  line-height: 1.4;
}

.config-main {
  flex: 1;
  height: 100%;
  display: flex;
  flex-direction: column;
  border: 1px solid var(--art-gray-300);
  border-radius: 8px;
  background: var(--default-box-color);
  overflow: hidden;
}

.config-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  border-bottom: 1px solid var(--art-gray-300);

  h3 {
    font-size: 16px;
    font-weight: 600;
    color: var(--art-gray-800);
    margin: 0 0 8px;
  }
}

.config-desc {
  font-size: 13px;
  color: var(--art-gray-500);
  margin: 0;
}

.config-content {
  flex: 1;
  padding: 24px;
  overflow-y: auto;
}

.config-section {
  max-width: 720px;
}

.config-sub-tabs {
  margin-bottom: 20px;

  :deep(.el-tabs__header) {
    margin-bottom: 0;
  }
}

.config-form {
  :deep(.el-form-item) {
    margin-bottom: 20px;
  }
}

.config-form-item {
  &:hover {
    .config-menu-btn {
      opacity: 1;
    }
  }
}

.config-item-wrapper {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  width: 100%;
}

.config-item-content {
  flex: 1;
  min-width: 0;
}

.config-menu-btn {
  flex-shrink: 0;
  width: 28px;
  height: 28px;
  min-height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s;
  color: var(--art-gray-500);
  opacity: 0;
  margin-top: 2px;

  &:hover {
    background: var(--art-gray-200);
    color: var(--art-gray-800);
  }
}

// 针对不同类型控件的菜单按钮对齐
.config-form-item {
  .config-item-wrapper {
    // 开关类型，按钮居中对齐
    &:has(.el-switch) {
      align-items: center;
      
      .config-menu-btn {
        margin-top: 0;
      }
    }
    
    // 复选框类型，按钮居中对齐
    &:has(.el-checkbox) {
      align-items: center;
      
      .config-menu-btn {
        margin-top: 0;
      }
    }
    
    // 单选框组，按钮顶部对齐
    &:has(.el-radio-group) {
      .config-menu-btn {
        margin-top: 6px;
      }
    }

    // 颜色选择器，按钮居中对齐
    &:has(.art-color-picker) {
      align-items: center;
      
      .config-menu-btn {
        margin-top: 0;
      }
    }

    // 日期时间选择器，按钮顶部对齐
    &:has(.el-date-picker),
    &:has(.el-time-picker),
    &:has(.el-cascader) {
      .config-menu-btn {
        margin-top: 4px;
      }
    }

    // 图片/文件上传，按钮顶部对齐
    &:has(.art-image-upload),
    &:has(.art-file-upload) {
      align-items: flex-start;
      
      .config-menu-btn {
        margin-top: 4px;
      }
    }

    // 富文本编辑器，按钮顶部对齐
    &:has(.art-wang-editor) {
      align-items: flex-start;
      
      .config-menu-btn {
        margin-top: 4px;
      }
    }
  }
}

.menu-item-hint {
  margin-left: 4px;
  font-size: 11px;
  color: var(--art-gray-400);
}

// 暗黑模式
:deep(.dark) {
  .config-menu-btn {
    &:hover {
      background: var(--art-gray-700);
      color: var(--art-gray-200);
    }
  }

  .group-menu-btn {
    &:hover {
      background: var(--art-gray-700);
      color: var(--art-gray-200);
    }
  }
}

.upload-area {
  width: 100%;
  max-width: 400px;
}

.logo-preview {
  width: 200px;
  height: 80px;
  border: 1px solid var(--art-gray-300);
  border-radius: 4px;
}

.upload-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px;
  border: 1px dashed var(--art-gray-400);
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.3s;

  &:hover {
    border-color: var(--theme-color);
    background: var(--theme-color-alpha-5);
  }

  span {
    margin-top: 8px;
    font-size: 13px;
    color: var(--art-gray-500);
  }
}

.config-actions {
  margin-top: 32px;
  padding-top: 24px;
  border-top: 1px solid var(--art-gray-200);
}

.config-footer {
  padding: 16px 24px;
  border-top: 1px solid var(--art-gray-300);
  background: var(--default-box-color);
}

.form-hint {
  margin-top: 4px;
  font-size: 12px;
  color: var(--art-gray-500);
}

.form-item-hint {
  margin-top: 4px;
  font-size: 12px;
  color: var(--art-gray-500);
  line-height: 1.5;
}
</style>
