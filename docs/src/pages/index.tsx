import clsx from 'clsx';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import HomepageFeatures from '@site/src/components/HomepageFeatures';
import Heading from '@theme/Heading';
import React, { ReactElement, useEffect, useState } from 'react';

import styles from './index.module.css';

function TerminalSwitcher() {
  const languages = [
    { label: 'English (us)', icon: '🇺🇸' },
    { label: 'English (uk)', icon: '🇬🇧' },
    { label: '日本語 (mozc)', icon: '🇯🇵' },
    { label: '한국어 (hangul)', icon: '🇰🇷' },
    { label: 'Tiếng Việt (telex)', icon: '🇻🇳' },
    { label: 'ไทย (kedmanee)', icon: '🇹🇭' },
    { label: 'Français (azerty)', icon: '🇫🇷' },
    { label: 'Deutsch (qwertz)', icon: '🇩🇪' },
    { label: 'Español (es)', icon: '🇪🇸' },
    { label: 'Português (br)', icon: '🇧🇷' },
    { label: 'Italiano (it)', icon: '🇮🇹' },
    { label: 'Русский (ru)', icon: '🇷🇺' },
    { label: 'Українська (ua)', icon: '🇺🇦' },
    { label: 'Polski (pl)', icon: '🇵🇱' },
    { label: 'Türkçe (tr)', icon: '🇹🇷' },
    { label: 'Ελληνικά (el)', icon: '🇬🇷' },
    { label: 'Nederlands (nl)', icon: '🇳🇱' },
    { label: 'Magyar (hu)', icon: '🇭🇺' },
    { label: 'Čeština (cz)', icon: '🇨🇿' },
    { label: 'Română (ro)', icon: '🇷🇴' },
    { label: 'עברית (he)', icon: '🇮🇱' },
    { label: 'العربية (ar)', icon: '🇸🇦' },
    { label: 'हिन्दी (hi)', icon: '🇮🇳' },
    { label: '中文 (rime)', icon: '🇨🇳' },
  ];
  const [index, setIndex] = useState(0);

  useEffect(() => {
    const timer = setTimeout(() => {
      setIndex((prev) => (prev + 1) % languages.length);
    }, 1400);
    return () => clearTimeout(timer);
  }, [index, languages.length]);

  return (
    <span className={styles.output}>
      {languages[index].icon} Switching to: {languages[index].label}
    </span>
  );
}

function HomepageHeader() {
  const {siteConfig} = useDocusaurusContext();
  return (
    <header className={clsx('hero hero--primary', styles.heroBanner)}>
      <div className="container">
        <div className={styles.heroContent}>
          <div className={styles.heroText}>
            <Heading as="h1" className="hero__title">
              {siteConfig.title}
            </Heading>
            <p className="hero__subtitle">{siteConfig.tagline}</p>
            <div className={styles.heroDescription}>
              <p>
                A smart, performance-focused input method switcher built for <strong>Hyprland</strong>.
                Automatically switches input methods based on active applications with zero configuration.
              </p>
            </div>
            <div className={styles.buttons}>
              <Link
                className="button button--secondary button--lg"
                to="/docs/intro">
                🚀 Get Started
              </Link>
              <Link
                className="button button--secondary button--lg"
                to="https://github.com/icyleaf/hypr-input-switcher/releases"
                target="_blank"
                rel="noopener noreferrer">
                📦 Download
              </Link>
            </div>
          </div>
          <div className={styles.heroDemo}>
            <div className={styles.terminalWindow}>
              <div className={styles.terminalHeader}>
                <div className={styles.terminalButtons}>
                  <span className={clsx(styles.terminalButton, styles.red)}></span>
                  <span className={clsx(styles.terminalButton, styles.yellow)}></span>
                  <span className={clsx(styles.terminalButton, styles.green)}></span>
                </div>
                <div className={styles.terminalTitle}>terminal</div>
              </div>
              <div className={styles.terminalBody}>
                <div className={styles.terminalLine}>
                  <span className={styles.prompt}>$</span> hypr-input-switcher --watch
                </div>
                <div className={styles.terminalLine}>
                  <span className={styles.output}>🎯 Detected window: firefox</span>
                </div>
                <div className={styles.terminalLine}>
                  <TerminalSwitcher />
                </div>
                <div className={styles.terminalLine}>
                  <span className={styles.output}>✅ Input method switched successfully</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </header>
  );
}

function QuickStartSection() {
  return (
    <section className={styles.quickStart}>
      <div className="container">
        <div className="row">
          <div className="col col--12">
            <Heading as="h2" className={styles.sectionTitle}>
              ⚡ Quick Start
            </Heading>
            <p className={styles.sectionSubtitle}>
              Get up and running in less than 30 seconds
            </p>
          </div>
        </div>
        <div className="row">
          <div className="col col--4">
            <div className={styles.quickStartStep}>
              <div className={styles.stepNumber}>1</div>
              <h3>📦 Install</h3>
              <div className={styles.codeBlock}>
                <code>paru -S hypr-input-switcher-bin</code>
              </div>
              <p>Or download from GitHub releases</p>
            </div>
          </div>
          <div className="col col--4">
            <div className={styles.quickStartStep}>
              <div className={styles.stepNumber}>2</div>
              <h3>⚙️ Configure</h3>
              <div className={styles.codeBlock}>
                <code>exec-once = hypr-input-switcher</code>
              </div>
              <p>Add to your Hyprland config</p>
            </div>
          </div>
          <div className="col col--4">
            <div className={styles.quickStartStep}>
              <div className={styles.stepNumber}>3</div>
              <h3>🎉 Enjoy</h3>
              <div className={styles.codeBlock}>
                <code>✨ Automatic switching!</code>
              </div>
              <p>Input methods switch automatically</p>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}

function StatsSection() {
  const stats = [
    { label: 'Zero-config setup', value: '< 30s', icon: '⚡' },
    { label: 'Memory usage', value: '< 5MB', icon: '🏃' },
    { label: 'Switch latency', value: '< 100ms', icon: '⚡' },
    { label: 'Supported apps', value: '∞', icon: '🎯' },
  ];

  return (
    <section className={styles.stats}>
      <div className="container">
        <div className="row">
          <div className="col col--12">
            <Heading as="h2" className={styles.sectionTitle}>
              🔥 Performance First
            </Heading>
          </div>
        </div>
        <div className="row">
          {stats.map((stat, idx) => (
            <div key={idx} className="col col--3">
              <div className={styles.statCard}>
                <div className={styles.statIcon}>{stat.icon}</div>
                <div className={styles.statValue}>{stat.value}</div>
                <div className={styles.statLabel}>{stat.label}</div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

export default function Home(): ReactElement {
  const {siteConfig} = useDocusaurusContext();
  return (
    <Layout
      title={siteConfig.title}
      description="Smart input method switcher for Hyprland - automatic, fast, and configurable">
      <HomepageHeader />
      <main>
        <HomepageFeatures />
        <QuickStartSection />
        <StatsSection />
      </main>
    </Layout>
  );
}
