export function getApiErrorMessage(error: unknown, fallback: string): string {
  const apiError = error as {
    response?: { data?: { error?: { message?: string } } }
    message?: string
  }

  const message = apiError.response?.data?.error?.message
  if (message) return message
  if (apiError.message === 'Network Error') {
    return 'Сервис ИИ недоступен. Проверьте подключение сервера и настройки OpenAI.'
  }
  return fallback
}
