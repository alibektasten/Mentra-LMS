'use client'

import { useEffect, useState } from 'react'
import { Bell, Check, Globe2, Loader2, Moon, ShieldCheck, Upload, UserRound } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { useMe, useUpdateProfile } from '@/lib/queries/users.queries'
import { useAuthStore } from '@/lib/stores/auth.store'
import { getApiErrorMessage } from '@/lib/api/error-message'
import { useI18n } from '@/lib/i18n'
import { publicAvatarURL, uploadAvatar } from '@/lib/api/private-files'

type Preferences = { language: 'ru' | 'en' | 'kk'; emailUpdates: boolean; deadlineAlerts: boolean; compactMode: boolean; theme: 'light' | 'dark'; uiSize: 'small' | 'medium' | 'large' }
const defaults: Preferences = { language: 'ru', emailUpdates: true, deadlineAlerts: true, compactMode: false, theme: 'light', uiSize: 'medium' }

function Toggle({ checked, onChange, label, description }: { checked: boolean; onChange: (value: boolean) => void; label: string; description: string }) {
  return <label className="flex cursor-pointer items-start justify-between gap-5 py-4">
    <span><span className="block text-sm font-semibold text-[#25251f]">{label}</span><span className="mt-1 block text-xs leading-relaxed text-[#85847d]">{description}</span></span>
    <button type="button" role="switch" aria-checked={checked} onClick={() => onChange(!checked)} className={`mt-1 h-6 w-11 rounded-full p-0.5 transition-colors ${checked ? 'bg-emerald-600' : 'bg-[#d9d8d2]'}`}><span className={`block h-5 w-5 rounded-full bg-white shadow-sm transition-transform ${checked ? 'translate-x-5' : 'translate-x-0'}`} /></button>
  </label>
}

export default function PersonalSettingsPage() {
  const { data: profile, isLoading } = useMe()
  const updateProfile = useUpdateProfile()
  const { user, role } = useAuthStore()
  const { t } = useI18n()
  const [name, setName] = useState('')
  const [avatarURL, setAvatarURL] = useState('')
  const [avatarSourceURL, setAvatarSourceURL] = useState('')
  const [avatarPreviewURL, setAvatarPreviewURL] = useState('')
  const [uploadingAvatar, setUploadingAvatar] = useState(false)
  const [preferences, setPreferences] = useState<Preferences>(defaults)

  useEffect(() => { if (profile) { setName(profile.name); setAvatarURL(profile.avatar_url ?? ''); setAvatarSourceURL(profile.avatar_source_url ?? '') } }, [profile])
  useEffect(() => {
    if (!user?.id) return
    try { setPreferences({ ...defaults, ...JSON.parse(localStorage.getItem(`mentra:preferences:${user.id}`) ?? '{}') }) } catch { setPreferences(defaults) }
  }, [user?.id])

  const savePreferences = (next: Preferences) => {
    setPreferences(next)
    document.documentElement.dataset.theme = next.theme
    document.documentElement.dataset.uiSize = next.uiSize
    document.documentElement.lang = next.language
    if (user?.id) localStorage.setItem(`mentra:preferences:${user.id}`, JSON.stringify(next))
    window.dispatchEvent(new Event('mentra:preferences'))
    toast.success('Настройки сохранены на этом устройстве')
  }
  const saveProfile = () => {
    const value = name.trim()
    if (!value) return toast.error('Укажите имя')
    updateProfile.mutate({ name: value, avatar_url: avatarSourceURL.trim() }, { onSuccess: () => { setAvatarPreviewURL(''); toast.success('Профиль обновлён') }, onError: (error) => toast.error(getApiErrorMessage(error, 'Не удалось обновить профиль')) })
  }
  const handleAvatarUpload = async (file?: File) => {
    if (!file) return
    if (!file.type.startsWith('image/') || file.size > 5 * 1024 * 1024) return toast.error('Выберите изображение до 5 МБ')
    setUploadingAvatar(true)
    try {
      const uploaded = await uploadAvatar(file)
      setAvatarURL(publicAvatarURL(user?.id ?? '', uploaded.path))
      setAvatarSourceURL(uploaded.path)
      setAvatarPreviewURL(URL.createObjectURL(file))
      toast.success('Фото профиля обновлено')
    } catch (error) { toast.error(getApiErrorMessage(error, 'Не удалось загрузить фото')) }
    finally { setUploadingAvatar(false) }
  }

  return <div className="mx-auto max-w-3xl space-y-6 pb-10">
    <div><p className="text-[11px] font-semibold uppercase tracking-[0.14em] text-emerald-700">{t('account')}</p><h1 className="mt-1 text-3xl font-bold tracking-[-0.04em] text-[#1d1d19]">{t('personalSettings')}</h1><p className="mt-2 text-sm text-[#77766f]">Профиль и предпочтения, доступные для роли {role ?? 'user'}.</p></div>
    <section className="overflow-hidden rounded-2xl border border-[#e5e4df] bg-white"><div className="flex items-center gap-3 border-b border-[#ecebe6] bg-[#fbfbf9] px-5 py-4"><UserRound className="h-5 w-5 text-emerald-700" /><div><h2 className="text-sm font-bold text-[#25251f]">Профиль</h2><p className="text-xs text-[#85847d]">Основные данные аккаунта</p></div></div><div className="space-y-4 p-5"><label className="block text-xs font-semibold text-[#5f5e57]">Отображаемое имя<Input className="mt-2" value={name} disabled={isLoading} onChange={(event) => setName(event.target.value)} /></label><div><p className="text-xs font-semibold text-[#5f5e57]">Фото профиля</p><div className="mt-2 flex items-center gap-3">{avatarURL ? <img src={avatarPreviewURL || avatarURL} alt="Фото профиля" className="h-14 w-14 rounded-full border object-cover" /> : <div className="flex h-14 w-14 items-center justify-center rounded-full bg-emerald-100 text-emerald-700"><UserRound className="h-6 w-6" /></div>}<label className="inline-flex cursor-pointer items-center gap-2 rounded-lg border border-[#d9d8d2] px-3 py-2 text-xs font-semibold"><Upload className="h-3.5 w-3.5" />{uploadingAvatar ? 'Загрузка…' : 'Выбрать фото'}<input type="file" accept="image/png,image/jpeg,image/webp" className="hidden" disabled={uploadingAvatar} onChange={(event) => handleAvatarUpload(event.target.files?.[0])} /></label></div></div><div className="rounded-xl bg-[#f6f6f3] px-4 py-3"><p className="text-[11px] font-semibold uppercase tracking-wide text-[#85847d]">Email</p><p className="mt-1 text-sm text-[#3d3c37]">{profile?.email ?? '—'}</p></div><Button onClick={saveProfile} disabled={updateProfile.isPending || isLoading || uploadingAvatar} className="bg-emerald-600 hover:bg-emerald-700">{updateProfile.isPending ? <><Loader2 className="h-4 w-4 animate-spin" /> Сохраняем</> : <><Check className="h-4 w-4" /> Сохранить профиль</>}</Button></div></section>
    <section className="overflow-hidden rounded-2xl border border-[#e5e4df] bg-white"><div className="flex items-center gap-3 border-b border-[#ecebe6] bg-[#fbfbf9] px-5 py-4"><Globe2 className="h-5 w-5 text-sky-700" /><div><h2 className="text-sm font-bold text-[#25251f]">Интерфейс</h2><p className="text-xs text-[#85847d]">Язык, тема и размер для этого аккаунта</p></div></div><div className="divide-y divide-[#ecebe6] px-5"><label className="flex items-center justify-between gap-4 py-4"><span><span className="block text-sm font-semibold text-[#25251f]">Язык</span><span className="mt-1 block text-xs text-[#85847d]">Для персонального интерфейса</span></span><select value={preferences.language} onChange={(event) => savePreferences({ ...preferences, language: event.target.value as Preferences['language'] })} className="rounded-lg border border-[#dfded8] bg-white px-3 py-2 text-sm"><option value="ru">Русский</option><option value="kk">Қазақша</option><option value="en">English</option></select></label><label className="flex items-center justify-between gap-4 py-4"><span><span className="block text-sm font-semibold text-[#25251f]">Тема</span><span className="mt-1 block text-xs text-[#85847d]">Применяется сразу ко всему рабочему пространству</span></span><select value={preferences.theme} onChange={(event) => savePreferences({ ...preferences, theme: event.target.value as Preferences['theme'] })} className="rounded-lg border border-[#dfded8] bg-white px-3 py-2 text-sm"><option value="light">Светлая</option><option value="dark">Тёмная</option></select></label><label className="flex items-center justify-between gap-4 py-4"><span><span className="block text-sm font-semibold text-[#25251f]">Размер интерфейса</span><span className="mt-1 block text-xs text-[#85847d]">Размер текста и элементов страницы</span></span><select value={preferences.uiSize} onChange={(event) => savePreferences({ ...preferences, uiSize: event.target.value as Preferences['uiSize'] })} className="rounded-lg border border-[#dfded8] bg-white px-3 py-2 text-sm"><option value="small">Маленький</option><option value="medium">Стандартный</option><option value="large">Крупный</option></select></label><Toggle checked={preferences.compactMode} onChange={(compactMode) => savePreferences({ ...preferences, compactMode })} label="Компактный режим" description="Более плотное отображение списков и карточек." /></div></section>
    <section className="overflow-hidden rounded-2xl border border-[#e5e4df] bg-white"><div className="flex items-center gap-3 border-b border-[#ecebe6] bg-[#fbfbf9] px-5 py-4"><Bell className="h-5 w-5 text-amber-600" /><div><h2 className="text-sm font-bold text-[#25251f]">Уведомления</h2><p className="text-xs text-[#85847d]">Контролируйте напоминания в приложении</p></div></div><div className="divide-y divide-[#ecebe6] px-5"><Toggle checked={preferences.deadlineAlerts} onChange={(deadlineAlerts) => savePreferences({ ...preferences, deadlineAlerts })} label="Сроки и дедлайны" description="Напоминать о ближайших заданиях, экзаменах и проверках." /><Toggle checked={preferences.emailUpdates} onChange={(emailUpdates) => savePreferences({ ...preferences, emailUpdates })} label="Email-обновления" description="Получать важные сообщения об активности в обучении." /></div></section>
    <section className="rounded-2xl border border-[#e5e4df] bg-[#fbfbf9] p-5"><div className="flex gap-3"><ShieldCheck className="mt-0.5 h-5 w-5 shrink-0 text-emerald-700" /><div><h2 className="text-sm font-bold text-[#25251f]">Безопасность аккаунта</h2><p className="mt-1 text-xs leading-relaxed text-[#77766f]">Вход через Google и пароль управляются на странице входа. Email и роль изменяются только администратором организации.</p></div></div></section>
  </div>
}
