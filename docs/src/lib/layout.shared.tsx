import type { BaseLayoutProps } from 'fumadocs-ui/layouts/shared';
import Image from 'next/image';

// fill this with your actual GitHub info, for example:
export const gitConfig = {
  user: 'pupload',
  repo: 'pupload',
  branch: 'main',
};

export function baseOptions(): BaseLayoutProps {
  return {
    nav: {
      title: (
          <Image src={"/pupload-text.svg"} alt={"Pupload logo"} width={120} height={120}/>
      )
    },
    githubUrl: `https://github.com/${gitConfig.user}/${gitConfig.repo}`,
  };
}
