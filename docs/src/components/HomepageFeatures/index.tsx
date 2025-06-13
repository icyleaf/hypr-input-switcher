import clsx from 'clsx';
import Heading from '@theme/Heading';
import { ReactElement } from 'react';
import React from 'react';
import styles from './styles.module.css';

type FeatureItem = {
  title: string;
  emoji: string;
  description: ReactElement;
};

const FeatureList: FeatureItem[] = [
  {
    title: 'Zero Configuration',
    emoji: '🚀',
    description: (
      <>
        Works out of the box with intelligent defaults.
        No complex setup required - just install and run.
      </>
    ),
  },
  {
    title: 'Lightning Fast',
    emoji: '⚡',
    description: (
      <>
        Built with Go for maximum performance.
        Switches input methods in under 100ms with minimal memory usage.
      </>
    ),
  },
  {
    title: 'Smart Detection',
    emoji: '🎯',
    description: (
      <>
        Automatically detects application context and switches
        to the appropriate input method based on your patterns.
      </>
    ),
  },
  {
    title: 'Highly Configurable',
    emoji: '🔧',
    description: (
      <>
        YAML-based configuration with hot-reload support.
        Customize rules, notifications, and behavior to your needs.
      </>
    ),
  },
  {
    title: 'Rich Notifications',
    emoji: '🔔',
    description: (
      <>
        Multiple notification backends with emoji support.
        Visual feedback for every input method switch.
      </>
    ),
  },
  {
    title: 'Developer Friendly',
    emoji: '🛠',
    description: (
      <>
        Comprehensive logging, debugging tools, and extensive
        documentation for easy troubleshooting and customization.
      </>
    ),
  },
];

function Feature({title, emoji, description}: FeatureItem) {
  return (
    <div className={clsx('col col--4')}>
      <div className={styles.featureCard}>
        <div className={styles.featureIcon}>
          <span>{emoji}</span>
        </div>
        <div className={styles.featureContent}>
          <Heading as="h3">{title}</Heading>
          <p>{description}</p>
        </div>
      </div>
    </div>
  );
}

export default function HomepageFeatures(): ReactElement {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className="row">
          <div className="col col--12">
            <Heading as="h2" className={styles.sectionTitle}>
              ✨ Why Choose Hypr Input Switcher?
            </Heading>
            <p className={styles.sectionSubtitle}>
              Built specifically for Hyprland with modern performance and usability in mind
            </p>
          </div>
        </div>
        <div className="row">
          {FeatureList.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}
