import { ref } from 'vue'

export function useFormValidation() {
  const errors = ref({})

  // 验证RSS URL格式
  const validateRssUrl = (url) => {
    if (!url || url.trim() === '') {
      return '请输入URL'
    }
    try {
      new URL(url)
      return ''
    } catch {
      return '请输入有效的URL'
    }
  }

  // 验证RSS源名称
  const validateRssName = (name) => {
    if (!name || name.trim() === '') {
      return '请输入源名称'
    }
    if (name.length > 100) {
      return '源名称不能超过100个字符'
    }
    return ''
  }

  // 验证频率
  const validateFrequency = (frequency) => {
    const validFrequencies = ['hourly', '30min', '2hours', 'daily']
    if (!validFrequencies.includes(frequency)) {
      return '请选择有效的更新频率'
    }
    return ''
  }

  // 验证模型名称
  const validateModelName = (name) => {
    if (!name || name.trim() === '') {
      return '请输入模型名称'
    }
    if (name.length > 100) {
      return '模型名称不能超过100个字符'
    }
    return ''
  }

  // 验证服务商
  const validateProvider = (provider) => {
    const validProviders = ['OpenAI', 'Anthropic', 'Meta', 'Google']
    if (!validProviders.includes(provider)) {
      return '请选择有效的服务商'
    }
    return ''
  }

  // 验证API密钥
  const validateApiKey = (apiKey) => {
    if (!apiKey || apiKey.trim() === '') {
      return '请输入API密钥'
    }
    if (apiKey.length < 10) {
      return 'API密钥长度至少为10个字符'
    }
    return ''
  }

  // 验证RSS表单
  const validateRssForm = (formData) => {
    errors.value = {}

    const nameError = validateRssName(formData.name)
    if (nameError) errors.value.name = nameError

    const urlError = validateRssUrl(formData.url)
    if (urlError) errors.value.url = urlError

    const frequencyError = validateFrequency(formData.frequency)
    if (frequencyError) errors.value.frequency = frequencyError

    return Object.keys(errors.value).length === 0
  }

  // 验证模型表单
  const validateModelForm = (formData) => {
    errors.value = {}

    const nameError = validateModelName(formData.name)
    if (nameError) errors.value.name = nameError

    const providerError = validateProvider(formData.provider)
    if (providerError) errors.value.provider = providerError

    const apiKeyError = validateApiKey(formData.apiKey)
    if (apiKeyError) errors.value.apiKey = apiKeyError

    return Object.keys(errors.value).length === 0
  }

  // 清除错误
  const clearErrors = () => {
    errors.value = {}
  }

  return {
    errors,
    validateRssUrl,
    validateRssName,
    validateFrequency,
    validateModelName,
    validateProvider,
    validateApiKey,
    validateRssForm,
    validateModelForm,
    clearErrors,
  }
}
