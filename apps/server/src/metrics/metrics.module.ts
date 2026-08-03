import { Module } from '@nestjs/common';
import { HostMetricsService } from './host-metrics.service';

/**
 * Holds agent-reported resource usage. Depends on nothing, so both the agents
 * gateway (writer) and the hosts module (reader) can import it without cycles.
 */
@Module({
  providers: [HostMetricsService],
  exports: [HostMetricsService],
})
export class MetricsModule {}
