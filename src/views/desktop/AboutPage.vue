<template>
    <v-row class="match-height">
        <v-col cols="12">
            <v-card>
                <template #title>
                    <span>{{ tt('global.app.title') }}</span>
                </template>

                <v-card-text>
                    <v-row no-gutters>
                        <v-col cols="12" md="2">
                            <span class="text-body-1">{{ tt('Version') }}</span>
                        </v-col>
                        <v-col cols="12" md="10" class="mb-6">
                            <span class="text-body-1">{{ clientVersion }}</span>
                        </v-col>
                    </v-row>
                    <v-row no-gutters v-if="clientBuildTime">
                        <v-col cols="12" md="2">
                            <span class="text-body-1">{{ tt('Build Time') }}</span>
                        </v-col>
                        <v-col cols="12" md="10" class="mb-6">
                            <span class="text-body-1">{{ clientBuildTime }}</span>
                        </v-col>
                    </v-row>
                    <v-row no-gutters>
                        <v-col cols="12" md="2">
                            <span class="text-body-1">{{ tt('Official Website') }}</span>
                        </v-col>
                        <v-col cols="12" md="10" class="mb-6">
                            <a class="text-body-1" href="https://github.com/James-HoneyBadger/ratel-ledger" target="_blank">
                                https://github.com/James-HoneyBadger/ratel-ledger
                            </a>
                        </v-col>
                    </v-row>
                    <v-row no-gutters>
                        <v-col cols="12" md="2">
                            <span class="text-body-1">{{ tt('Report Issue') }}</span>
                        </v-col>
                        <v-col cols="12" md="10" class="mb-6">
                            <a class="text-body-1" href="https://github.com/James-HoneyBadger/ratel-ledger/issues" target="_blank">
                                https://github.com/James-HoneyBadger/ratel-ledger/issues
                            </a>
                        </v-col>
                    </v-row>
                    <v-row no-gutters>
                        <v-col cols="12" md="2">
                            <span class="text-body-1">{{ tt('Getting help') }}</span>
                        </v-col>
                        <v-col cols="12" md="10">
                            <a class="text-body-1" href="https://github.com/James-HoneyBadger/ratel-ledger" target="_blank">
                                https://github.com/James-HoneyBadger/ratel-ledger
                            </a>
                        </v-col>
                    </v-row>
                </v-card-text>
            </v-card>
        </v-col>

        <v-col cols="12" v-if="exchangeRatesData && !isUserCustomExchangeRates">
            <v-card :title="tt('Exchange Rates Data')">
                <v-card-text>
                    <v-row no-gutters>
                        <v-col cols="12" md="2">
                            <span class="text-body-1">{{ tt('Provider') }}</span>
                        </v-col>
                        <v-col cols="12" md="10">
                            <a class="text-body-1" :href="exchangeRatesData.referenceUrl" target="_blank"
                               v-if="exchangeRatesData.referenceUrl">{{ exchangeRatesData.dataSource }}</a>
                            <span class="text-body-1" v-if="!exchangeRatesData.referenceUrl">{{ exchangeRatesData.dataSource }}</span>
                        </v-col>
                    </v-row>
                </v-card-text>
            </v-card>
        </v-col>

        <v-col cols="12" v-if="mapProviderName">
            <v-card :title="tt('Map')">
                <v-card-text>
                    <v-row no-gutters>
                        <v-col cols="12" md="2">
                            <span class="text-body-1">{{ tt('Provider') }}</span>
                        </v-col>
                        <v-col cols="12" md="10">
                            <a class="text-body-1" :href="mapProviderWebsite" target="_blank"
                               v-if="mapProviderWebsite">{{ mapProviderName }}</a>
                            <span class="text-body-1" v-if="!mapProviderWebsite">{{ mapProviderName }}</span>
                        </v-col>
                    </v-row>
                </v-card-text>
            </v-card>
        </v-col>

        <v-col cols="12">
            <v-card :title="tt('License')">
                <v-card-text>
                    <v-row no-gutters>
                        <v-col cols="12">
                            <p>
                                <span :key="num" v-for="(line, num) in licenseLines"
                                      :style="{ 'display': line ? 'initial' : 'block', 'padding' : line ? '0' : '0 0 1em 0' }">
                                    {{ line }}
                                </span>
                            </p>
                            <v-divider/>
                            <br/>
                            <p>
                                <span>Copyright © 2020–2026 Honey Badger Universe</span>
                            </p>
                            <p class="mt-4 mb-4">
                                <span>Honey Ledger also contains additional third party software and illustration.</span><br/>
                                <span>All the third party software / illustration included or linked is redistributed under the terms and conditions of their original licenses.</span>
                            </p>
                            <p :key="licenseInfo.name" v-for="licenseInfo in thirdPartyLicenses">
                                <strong>{{ licenseInfo.name }}</strong>
                                <br v-if="licenseInfo.copyright"/><span v-if="licenseInfo.copyright">{{ licenseInfo.copyright }}</span>
                                <br v-if="licenseInfo.licenseUrl"/><span class="work-break-all" v-if="licenseInfo.licenseUrl">{{ licenseInfo.license || 'License' }}: </span>
                                <a target="_blank" :href="licenseInfo.licenseUrl">{{ licenseInfo.licenseUrl }}</a>
                                <br v-if="licenseInfo.url"/><a class="work-break-all" target="_blank" :href="licenseInfo.url" v-if="licenseInfo.url">{{ licenseInfo.url }}</a>
                            </p>
                        </v-col>
                    </v-row>
                </v-card-text>
            </v-card>
        </v-col>
    </v-row>
</template>

<script setup lang="ts">
import { useI18n } from '@/locales/helpers.ts';
import { useAboutPageBase } from '@/views/base/AboutPageBase.ts';

const { tt } = useI18n();
const {
    clientVersion,
    clientBuildTime,
    exchangeRatesData,
    isUserCustomExchangeRates,
    mapProviderName,
    mapProviderWebsite,
    licenseLines,
    thirdPartyLicenses,
    init
} = useAboutPageBase();

init();
</script>

<style>
</style>
