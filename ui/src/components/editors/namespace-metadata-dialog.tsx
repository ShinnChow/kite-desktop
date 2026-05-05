import { useEffect, useState, type FormEvent } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import type { Namespace } from 'kubernetes-types/core/v1'
import { Plus, Tags, Trash2 } from 'lucide-react'
import { useTranslation } from 'react-i18next'
import { toast } from 'sonner'

import { updateResource } from '@/lib/api'
import { translateError } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

type MetadataType = 'labels' | 'annotations'

type KeyValueItem = {
  key: string
  value: string
}

function toKeyValueItems(items?: Record<string, string>) {
  return Object.entries(items || {}).map(([key, value]) => ({ key, value }))
}

function toMetadataRecord(items: KeyValueItem[]) {
  return items.reduce<Record<string, string>>((result, item) => {
    const key = item.key.trim()
    if (!key) {
      return result
    }

    result[key] = item.value
    return result
  }, {})
}

export function NamespaceMetadataDialog(props: {
  open: boolean
  onOpenChange: (open: boolean) => void
  namespace: Namespace | null
  type: MetadataType
}) {
  const { open, onOpenChange, namespace, type } = props
  const { t } = useTranslation()
  const queryClient = useQueryClient()
  const [items, setItems] = useState<KeyValueItem[]>([])
  const [isSaving, setIsSaving] = useState(false)

  useEffect(() => {
    if (!open || !namespace) {
      return
    }

    const source =
      type === 'labels'
        ? namespace.metadata?.labels
        : namespace.metadata?.annotations
    setItems(toKeyValueItems(source))
    setIsSaving(false)
  }, [namespace, open, type])

  const handleItemChange = (
    index: number,
    field: keyof KeyValueItem,
    value: string
  ) => {
    setItems((current) =>
      current.map((item, itemIndex) =>
        itemIndex === index ? { ...item, [field]: value } : item
      )
    )
  }

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    const namespaceName = namespace?.metadata?.name
    if (!namespace || !namespaceName) {
      return
    }

    setIsSaving(true)

    try {
      const nextMetadata = toMetadataRecord(items)
      await updateResource('namespaces', namespaceName, undefined, {
        ...namespace,
        metadata: {
          ...(namespace.metadata || {}),
          [type]: nextMetadata,
        },
      })

      await queryClient.invalidateQueries({ queryKey: ['namespaces'] })
      toast.success(
        t('namespaceMetadataDialog.success', {
          defaultValue: 'Namespace metadata updated successfully',
        })
      )
      onOpenChange(false)
    } catch (error) {
      toast.error(translateError(error, t))
    } finally {
      setIsSaving(false)
    }
  }

  const titleKey =
    type === 'labels'
      ? 'namespaceMetadataDialog.labelsTitle'
      : 'namespaceMetadataDialog.annotationsTitle'

  const descriptionKey =
    type === 'labels'
      ? 'namespaceMetadataDialog.labelsDescription'
      : 'namespaceMetadataDialog.annotationsDescription'

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="flex h-[85vh] max-h-[85vh] flex-col overflow-hidden sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Tags className="h-4 w-4" />
            {t(titleKey, {
              defaultValue:
                type === 'labels' ? 'Manage labels' : 'Manage annotations',
            })}
          </DialogTitle>
          <DialogDescription>
            {t(descriptionKey, {
              defaultValue:
                type === 'labels'
                  ? 'Add or remove namespace labels.'
                  : 'Add or remove namespace annotations.',
            })}
          </DialogDescription>
        </DialogHeader>

        <form
          className="flex min-h-0 flex-1 flex-col overflow-hidden"
          onSubmit={handleSubmit}
        >
          <div className="flex items-center justify-between gap-3 pb-4">
            <Label>{t(`detail.fields.${type}`)}</Label>
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={() =>
                setItems((current) => [{ key: '', value: '' }, ...current])
              }
            >
              <Plus className="h-4 w-4" />
              {t('common.add', 'Add')}
            </Button>
          </div>

          <div className="min-h-0 flex-1 overflow-y-auto pr-1">
            {items.length === 0 ? (
              <p className="text-sm text-muted-foreground">
                {t('namespaceMetadataDialog.empty', {
                  defaultValue: 'No entries yet',
                })}
              </p>
            ) : null}

            <div className="space-y-3">
              {items.map((item, index) => (
                <div
                  key={`${type}-${index}`}
                  className="grid grid-cols-[1fr_1fr_auto] gap-2"
                >
                  <Input
                    value={item.key}
                    onChange={(event) =>
                      handleItemChange(index, 'key', event.target.value)
                    }
                    placeholder={t('common.key', 'Key')}
                  />
                  <Input
                    value={item.value}
                    onChange={(event) =>
                      handleItemChange(index, 'value', event.target.value)
                    }
                    placeholder={t('common.value', 'Value')}
                  />
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    aria-label={t('common.remove', 'Remove')}
                    onClick={() =>
                      setItems((current) =>
                        current.filter((_, itemIndex) => itemIndex !== index)
                      )
                    }
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </div>
              ))}
            </div>
          </div>

          <DialogFooter className="mt-4 shrink-0 border-t pt-4">
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={isSaving}
            >
              {t('common.cancel')}
            </Button>
            <Button type="submit" disabled={isSaving}>
              {isSaving
                ? t('common.saving', 'Saving')
                : t('common.save', 'Save')}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
