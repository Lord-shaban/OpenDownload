"use client";
import { GlobeSimpleIcon } from "@phosphor-icons/react/dist/csr/GlobeSimple";
import { CaretDownIcon } from "@phosphor-icons/react/dist/csr/CaretDown";
import { InstagramLogoIcon } from "@phosphor-icons/react/dist/csr/InstagramLogo";
import { TiktokLogoIcon } from "@phosphor-icons/react/dist/csr/TiktokLogo";
import { XLogoIcon } from "@phosphor-icons/react/dist/csr/XLogo";
import { FacebookLogoIcon } from "@phosphor-icons/react/dist/csr/FacebookLogo";
import { LinkedinLogoIcon } from "@phosphor-icons/react/dist/csr/LinkedinLogo";
import { PinterestLogoIcon } from "@phosphor-icons/react/dist/csr/PinterestLogo";
import { ThreadsLogoIcon } from "@phosphor-icons/react/dist/csr/ThreadsLogo";
import { RedditLogoIcon } from "@phosphor-icons/react/dist/csr/RedditLogo";
import { SoundcloudLogoIcon } from "@phosphor-icons/react/dist/csr/SoundcloudLogo";
import { PlayCircleIcon } from "@phosphor-icons/react/dist/csr/PlayCircle";
import { useLocale } from "@/components/locale-provider";
import { sources } from "@/lib/media";

const logos: Record<string, typeof GlobeSimpleIcon> = {
  Instagram: InstagramLogoIcon,
  TikTok: TiktokLogoIcon,
  "X / Twitter": XLogoIcon,
  Facebook: FacebookLogoIcon,
  LinkedIn: LinkedinLogoIcon,
  Pinterest: PinterestLogoIcon,
  Threads: ThreadsLogoIcon,
  Reddit: RedditLogoIcon,
  Vimeo: PlayCircleIcon,
  SoundCloud: SoundcloudLogoIcon,
};

export function SupportedSites() {
  const { t } = useLocale();
  return (
    <details className="supported-sites">
      <summary>
        <GlobeSimpleIcon size={16} weight="duotone" aria-hidden="true" />
        {t("supportedSites")}
        <CaretDownIcon size={12} className="sites-chevron" aria-hidden="true" />
      </summary>
      <div className="sites-panel glass-panel panel-enter">
        <ul aria-label={t("supportedSites")}>
          {sources.map((source) => {
            const Icon = logos[source] || GlobeSimpleIcon;
            return (
              <li key={source}>
                <Icon size={20} aria-hidden="true" />
                <bdi>{source}</bdi>
              </li>
            );
          })}
        </ul>
        <div className="sites-note">
          <span>{t("directLinks")}</span>
          <p>{t("siteAvailability")}</p>
          <p>{t("youtubeUnavailableNote")}</p>
        </div>
      </div>
    </details>
  );
}
