"use client";
import { CheckIcon as Check } from "@phosphor-icons/react/dist/csr/Check";
import { useLocale } from "@/components/locale-provider";
export function FlowProgress({ step }: { step: 1 | 2 | 3 }) {
  const { t } = useLocale();
  return (
    <ol className="flow-progress" aria-label={t("downloadFlow")}>
      {(["addLinkStep", "chooseFormatStep", "saveStep"] as const).map(
        (key, index) => (
          <li
            key={key}
            aria-current={step === index + 1 ? "step" : undefined}
            data-done={step > index + 1}
          >
            <span className="flow-number" aria-hidden="true">
              {step > index + 1 ? <Check size={13} /> : index + 1}
            </span>
            <span>{t(key)}</span>
          </li>
        ),
      )}
    </ol>
  );
}
