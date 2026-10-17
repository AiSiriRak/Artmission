import { useState } from "react";
import { Button } from "@/components/ui/Button";

interface DownloadArtworkButtonProps {
  imageUrl?: string;
  fileName: string;
}

export default function DownloadArtworkButton({
  imageUrl,
  fileName,
}: DownloadArtworkButtonProps) {
  const [downloading, setDownloading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Fetch as a blob because the `download` attribute is ignored for cross-origin URLs.
  const handleDownload = async () => {
    if (!imageUrl || downloading) return;
    setDownloading(true);
    setError(null);
    try {
      const res = await fetch(imageUrl);
      if (!res.ok) throw new Error("Download failed");
      const blob = await res.blob();
      const ext = blob.type.split("/")[1]?.split("+")[0] || "png";
      const href = URL.createObjectURL(blob);
      const link = document.createElement("a");
      link.href = href;
      link.download = `${fileName}.${ext}`;
      document.body.appendChild(link);
      link.click();
      link.remove();
      URL.revokeObjectURL(href);
    } catch {
      // Fall back to opening the file when the host blocks cross-origin fetches.
      window.open(imageUrl, "_blank", "noopener,noreferrer");
      setError("Could not download directly, opened in a new tab instead.");
    } finally {
      setDownloading(false);
    }
  };

  return (
    <div className="mt-6 flex flex-col items-center gap-2">
      <Button
        type="button"
        variant="dark"
        onClick={handleDownload}
        disabled={downloading || !imageUrl}
      >
        <span
          role="img"
          aria-label="download icon"
          className="inline-block w-6 h-6 bg-white"
          style={{
            maskImage: "url('/icons/download.svg')",
            maskRepeat: "no-repeat",
            maskPosition: "center",
            maskSize: "contain",
          }}
        />
        {downloading ? "Downloading..." : "Download"}
      </Button>
      {!imageUrl && <p className="text-small text-neutral">Original artwork download is not available yet.</p>}
      {error && <p className="text-small text-red-500">{error}</p>}
    </div>
  );
}
