import { useVersionInfo } from '@/lib/api'
import { openURL } from '@/lib/desktop'

export function VersionInfo() {
  const { data: versionInfo } = useVersionInfo()

  if (!versionInfo) return null

  const handleCommitClick = (e: React.MouseEvent) => {
    e.preventDefault()
    e.stopPropagation()
    // GitHub repository URL - you can modify this to match your repository
    const repoUrl = 'https://github.com/kite-org/kite'
    const commitUrl = `${repoUrl}/commit/${versionInfo.commitId}`
    void openURL(commitUrl)
  }

  return (
    <div className="text-[10px] text-muted-foreground/60 font-mono leading-none">
      v{versionInfo.version.replace(/^v/, '')} •{' '}
      <button
        onClick={handleCommitClick}
        className="hover:text-primary/80 hover:underline transition-colors cursor-pointer"
        title={`View commit ${versionInfo.commitId} on GitHub`}
      >
        {versionInfo.commitId.slice(0, 7)}
      </button>
    </div>
  )
}
